# Debugging a Race Condition: `history -a` and Readline's Auto-Save

## The Symptom

`history -a /tmp/pear.txt` was supposed to append session commands to a file. It worked locally but failed on CodeCrafters — the output file had only the original content plus blank lines, with **none** of the new session commands.

```
Expected:                         Got:
echo grape strawberry orange      echo grape strawberry orange
echo pineapple apple              echo pineapple apple
echo strawberry blueberry pear    (empty)
echo apple pear strawberry        (empty)
echo strawberry grape pineapple
history -a /tmp/pear.txt
```

## The Root Cause: A Goroutine Race Condition

The `chzyer/readline` library processes input in a background goroutine (`ioloop`). When the user presses Enter, two things happen **in sequence within the ioloop goroutine**, but **concurrently with the main goroutine**:

```
ioloop goroutine                    main goroutine
─────────────────                   ──────────────
1. outchan <- data  ─────────────>  receives from outchan
2. history.New(data)  (file write)  executes the command
```

Step 1 uses an **unbuffered channel**, which means:
- The send blocks until the receiver reads
- Once the receiver reads, **both goroutines resume simultaneously**

So step 2 (saving to the history file) and the command execution happen **concurrently** with no synchronization between them.

### Why It Worked Locally

When piping input (`echo "cmd" | ./shell`), all input is buffered in the pipe. The timing worked out because:
- `echo` commands are extremely fast (just print and return)
- The ioloop had time to finish the file write before the next command started
- Even if there was a race, the local system was fast enough to mask it

### Why It Failed on CodeCrafters

The tester uses a PTY (pseudo-terminal) to interact with the shell, which can have different timing characteristics. When `history -a` executed, the ioloop goroutine **hadn't finished writing the previous commands to the history file yet**. So `createNewHistory` read an empty (or incomplete) file.

## The Concept: Race Conditions

A **race condition** occurs when the behavior of a program depends on the relative timing of concurrent operations, and the outcome is non-deterministic.

### Key Ingredients of a Race

1. **Shared state** — the history file on disk
2. **Concurrent access** — ioloop writes to it, main goroutine reads it
3. **No synchronization** — nothing guarantees the write finishes before the read

### Why Races Are Hard to Find

- They often work in development but fail in production (different timing)
- They can depend on CPU speed, OS scheduler, I/O latency, load
- They may pass 99 out of 100 runs, then fail once
- Adding print statements or debugger breakpoints can **change the timing** and make the race disappear (called a "Heisenbug")

## How to Debug Races Like This

### 1. Read the Error Output Carefully

The test showed the file had the original lines plus **empty lines** — not garbage, not partial data. This was a clue that the history file was being read as empty (`strings.Split("", "\n")` produces `[""]` — a slice with one empty string).

### 2. Trace the Data Flow

Map out exactly where data comes from and where it goes:

```
User types command
  → readline ioloop reads it
  → ioloop sends to outchan (main goroutine receives)
  → ioloop writes to history FILE        ← writer
  → main goroutine runs `history -a`
  → createNewHistory reads history FILE   ← reader
```

When you see a writer and reader accessing the same resource from different goroutines, that's a red flag.

### 3. Check for Synchronization

Ask: "What guarantees the write completes before the read?"

In this case: **nothing**. The channel synchronization only guaranteed the command text was delivered to the main goroutine. The file write happened after the channel send, with no further coordination.

### 4. Verify Your Assumptions

The original code assumed: "If readline auto-saves history, the file will have the commands when I read it." This assumption conflated **"the command was sent"** with **"the command was persisted to disk"** — two different events with no guaranteed ordering relative to the main goroutine.

## The Fix

Instead of reading the history file (which is written by another goroutine), maintain an in-memory `sessionHistory` slice in the main goroutine:

```go
var sessionHistory []string

// In the main loop, right after Readline():
line, err := rl.Readline()
sessionHistory = append(sessionHistory, line)

// In createNewHistory, use sessionHistory instead of reading the file:
newLines = append(newLines, sessionHistory[historyAppendOffset:]...)
```

This eliminates the race entirely because:
- The write (`append`) and read (`sessionHistory[offset:]`) happen in the **same goroutine**
- The append happens **before** any command execution
- No file I/O timing dependency

## General Lessons

1. **Don't read back what another goroutine writes** — if you need data, keep it in the goroutine that uses it
2. **Channel sends synchronize the send, not side effects after it** — code after `ch <- data` runs concurrently with the receiver
3. **"Works locally" means nothing for concurrency bugs** — always reason about what's guaranteed, not what happens to work
4. **Go's race detector (`go run -race`) can catch these** — run it during development to surface races early
5. **Unbuffered channels guarantee a handoff, not a happens-before for subsequent code** — the sender and receiver both proceed independently after the handoff

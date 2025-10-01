# Testing Cluely Backend

This guide walks you through testing the backend to verify everything works correctly.

## Prerequisites

Before testing, you need:

1. **Gemini API Key** (required for AI responses)
   - Get it from: https://aistudio.google.com/app/apikey
   - Copy the key (starts with `AI...`)

2. **Google Cloud Speech-to-Text** (optional for this test)
   - The test script doesn't need ASR since it sends pre-transcribed text
   - You can skip setting up GCP credentials for now

## Step 1: Set Environment Variables

### Windows (PowerShell)

```powershell
# Navigate to server directory
cd server

# Set your Gemini API key
$env:GEMINI_API_KEY = "YOUR_API_KEY_HERE"

# Optional: Choose a specific model (default is gemini-flash-lite-latest)
# $env:GEMINI_MODEL = "gemini-1.5-flash"
```

### Alternative: Create a `.env` file

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
# Edit .env with your actual GEMINI_API_KEY
```

**Note:** The Go server doesn't automatically load `.env` files. You'll need to export them manually or use a tool like `godotenv`.

## Step 2: Start the Server

```bash
# Build the server
go build -o bin/cluelyd ./cmd/cluelyd

# Run the server
./bin/cluelyd
```

You should see:
```
cluelyd listening :8080
[metrics] sessions=0 pcm(in=0 drop=0) asr(p=0 f=0) hints=0 followups=0 errors(asr=0 ans=0)
```

**Keep this terminal running!**

## Step 3: Run the Test Script

Open a **new terminal** and run:

```bash
cd server

# Build the test client
go build -o bin/test ./cmd/test

# Run the test
./bin/test
```

## What to Expect

The test script will:

1. **Connect** to your server at `ws://localhost:8080/ws`
2. **Send 4 test scenarios** with different transcripts and OCR contexts
3. **Display responses** from Gemini in real-time

### Successful Output Example

```
🧪 Cluely Backend Test
Connecting to: ws://localhost:8080/ws

✅ Connected to server

📝 Test 1: Simple transcript
📄 Final transcript: We can probably meet Friday to discuss the budget
💡 HINT: 'Probably' signals low commitment—confirm availability now
   ↳ Follow-up: Ask: 'Friday 2pm works—shall I send invite?'

📝 Test 2: Transcript with visual context
📄 Final transcript: So our Q4 revenue was down 8 percent but user engagement is way up
💡 HINT: Pivot to engagement growth story—higher LTV potential
   ↳ Follow-up: Ask: 'What's driving the MAU increase?'

📝 Test 3: Sales meeting scenario
📄 Final transcript: I'm not sure if we have the budget for this right now
💡 HINT: Budget objection—offer flexible payment or ROI proof
   ↳ Follow-up: Ask: 'What budget range were you thinking?'

📝 Test 4: Technical explanation
📄 Final transcript: I don't really understand how this architecture works
💡 HINT: They're confused—simplify the explanation now
   ↳ Follow-up: Use analogy: 'Like separate apps talking via messages'

✅ All tests sent! Check responses above.
💡 Tip: If you see hints, Gemini is working correctly!
```

### If You See Fallback Hints

If you see generic hints like "Confirm budget owner", it means:
- ❌ Your `GEMINI_API_KEY` is not set correctly
- ❌ Or the API key is invalid/expired

Check the server logs for errors:
```
[answer] GEMINI_API_KEY missing, using fallback
```
or
```
[answer] gemini status 401: {...}
```

## Step 4: Check Server Logs

In the server terminal, you should see:

```
[answer] success (JSON): 'Probably' signals low commitment—confirm availability now
[metrics] sessions=1 pcm(in=0 drop=0) asr(p=0 f=4) hints=4 followups=4 errors(asr=0 ans=0)
```

This confirms:
- ✅ 4 final transcripts processed
- ✅ 4 hints generated
- ✅ 4 follow-ups generated
- ✅ 0 errors

## Troubleshooting

### Error: "Failed to connect"
- Make sure the server is running (`./bin/cluelyd`)
- Check port 8080 isn't blocked by firewall
- Try: `netstat -an | findstr :8080` (should show LISTENING)

### Error: "gemini status 400"
- Your prompt might be too long (unlikely with test data)
- Check server logs for the full error message

### Error: "gemini status 401" or "403"
- Your API key is invalid
- Get a new key from https://aistudio.google.com/app/apikey
- Make sure you copied the entire key (starts with `AI...`)

### No hints appearing
- Check if rate limiting is working (1 hint per 1.5s)
- The test script has delays built in, so this shouldn't happen
- Check server logs for `[answer]` errors

## Advanced Testing

### Test with Custom Transcript

```bash
# Modify server/cmd/test/main.go
# Add your own test case:
sendJSON(ctx, c, map[string]any{
    "type":  "transcript",
    "text":  "Your custom transcript here",
    "final": true,
})
```

### Test with OCR Context

```bash
# Send visual context before transcript:
sendJSON(ctx, c, map[string]any{
    "type":  "frame_meta",
    "ocr":   []string{"Token1", "Token2", "Token3"},
    "first": true,  // Marks as first frame of segment
})

# Then send transcript
sendJSON(ctx, c, map[string]any{
    "type":  "transcript",
    "text":  "Your transcript",
    "final": true,
})

# Optionally send last frame
sendJSON(ctx, c, map[string]any{
    "type": "frame_meta",
    "ocr":  []string{"Different", "Tokens"},
    "last": true,
})
```

## Next Steps

Once the backend test passes:

1. **Test with real ASR**: Set up Google Cloud credentials and test with actual audio
2. **Test from visionOS client**: Build the Swift app and connect to this server
3. **Monitor metrics**: Watch the periodic metrics logs to track performance
4. **Check backpressure**: Send lots of audio data to test the ASR buffer

## Quick Reference

```bash
# Start server (terminal 1)
cd server
go build -o bin/cluelyd ./cmd/cluelyd
./bin/cluelyd

# Run test (terminal 2)
cd server
go build -o bin/test ./cmd/test
./bin/test

# Build dev WS client (terminal 2, alternative)
go build -o bin/wsdev ./cmd/wsdev
./bin/wsdev
```

## Success Criteria

Your backend is working correctly if:
- ✅ Test script connects without errors
- ✅ You see intelligent, context-aware hints (not fallback hints)
- ✅ Follow-up suggestions are specific and actionable
- ✅ OCR context is incorporated into hints (Test 2-4)
- ✅ Server logs show 0 errors
- ✅ Metrics show correct counts

If all criteria pass, your backend is ready for the visionOS client! 🎉

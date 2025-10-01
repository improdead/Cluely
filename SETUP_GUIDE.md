# 🚀 Cluely Backend Setup & Testing Guide

## What's New

This guide covers the recent enhancements to the Cluely backend that dramatically improve AI coaching capabilities.

## ✅ Recent Enhancements

### 1. Environment Configuration (`.env.example`)
- **Template with all environment variables**
- Detailed descriptions and links to get API keys
- Shows optional vs required vars
- Located at: `server/.env.example`

### 2. Dramatically Enhanced Gemini Prompt
- **Elite executive coach persona** with strategic insight
- **5 coaching principles**: Specific, Actionable, Strategic, Concise, Context-Aware
- **Rich examples** showing different scenarios:
  - Q4 revenue discussions
  - Technical confusion handling
  - Commitment signal detection
- **Uses OCR context** with first/last flags to provide:
  - "Screen at start" context
  - "Screen at end" context
- **Much more powerful responses** - references actual screen content and provides strategic guidance

### 3. Backend OCR Integration
- Stores `firstOCR` and `lastOCR` from `frame_meta` messages
- Passes all three contexts (`ocr`, `firstOCR`, `lastOCR`) to Gemini
- Prompt differentiates between start/end screen content

### 4. Test Script (`server/cmd/test/main.go`)
- Simulates a visionOS client
- Sends 4 test scenarios with transcripts + OCR
- Shows real-time Gemini responses
- **No ASR needed** - sends pre-transcribed text

### 5. Comprehensive Testing Documentation (`TESTING.md`)
- Complete step-by-step testing guide
- Troubleshooting section
- Success criteria checklist

---

## 🚀 Quick Start Guide

Follow these steps to get your backend up and running with AI-powered coaching.

### Step 1: Get Your Gemini API Key

1. Go to: **https://aistudio.google.com/app/apikey**
2. Click **"Create API Key"**
3. Copy the key (starts with `AI...`)

### Step 2: Set the API Key

#### Option A - Quick Test (PowerShell)
```powershell
cd server
$env:GEMINI_API_KEY = "YOUR_API_KEY_HERE"
```

#### Option B - Permanent (System Environment Variables)
1. Windows Search → **"Environment Variables"**
2. Add new user variable: `GEMINI_API_KEY` = `your-key`

### Step 3: Build & Run Server

```powershell
cd server

# Build the server
go build -o bin/cluelyd ./cmd/cluelyd

# Run the server
./bin/cluelyd
```

**You should see:**
```
cluelyd listening :8080
[metrics] sessions=0 pcm(in=0 drop=0) asr(p=0 f=0) hints=0 followups=0 errors(asr=0 ans=0)
```

**Keep this terminal running!**

### Step 4: Run Test Script (New Terminal)

```powershell
cd server

# Build the test client
go build -o bin/test ./cmd/test

# Run the test
./bin/test
```

---

## 📊 What You'll See If It Works

### Test Output (Terminal 2):
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

### Server Logs (Terminal 1):
```
[answer] success (JSON): 'Probably' signals low commitment—confirm availability now
[answer] success (JSON): Pivot to engagement growth story—higher LTV potential
[answer] success (JSON): Budget objection—offer flexible payment or ROI proof
[answer] success (JSON): They're confused—simplify the explanation now
[metrics] sessions=1 pcm(in=0 drop=0) asr(p=0 f=4) hints=4 followups=4 errors(asr=0 ans=0)
```

---

## ⚠️ If You See Fallback Hints

If responses are generic like **"Confirm budget owner"**, it means:
- ❌ Your `GEMINI_API_KEY` isn't set correctly
- ❌ Check server logs for: `[answer] GEMINI_API_KEY missing, using fallback`

**To fix:**
1. Verify you copied the entire API key
2. Make sure you set it in the correct terminal before running the server
3. Try setting it as a system environment variable

---

## 🎯 Success Criteria

Your backend is working if:
- ✅ Test connects without errors
- ✅ You see intelligent, context-aware hints (not generic fallback)
- ✅ Follow-ups are specific and actionable
- ✅ Test 2-4 show OCR context being used (e.g., "Pivot to engagement growth story" references the Q4/MAU screen data)
- ✅ Server logs show 0 errors
- ✅ Metrics show 4 hints, 4 followups

---

## 🔥 New Prompt Power

The enhanced prompt now makes Gemini:

### 1. Reference Screen Content
**Example:** "Pivot to engagement growth story"
- Uses OCR data: `$2.1M, MAU: 45K`
- Provides strategic pivot based on what's on screen

### 2. Detect Subtle Signals
**Example:** "'Probably' signals low commitment"
- Catches hesitation words
- Suggests immediate action to lock in commitment

### 3. Think Strategically
**Example:** "They're confused—simplify now"
- Recognizes when audience is lost
- Provides tactical intervention

### 4. Give Tactical Next Steps
**Example:** "Ask: 'Friday 2pm works—shall I send invite?'"
- Specific, actionable dialogue suggestions
- Moves conversation forward

### 5. Use Analogies
**Example:** "Like separate apps talking via messages"
- Simplifies technical concepts
- Makes complex ideas accessible

**This is way more powerful than the simple prompt before!**

---

## 📁 Key Files Reference

| File | Purpose |
|------|---------|
| `server/.env.example` | Environment variable template with descriptions |
| `TESTING.md` | Complete testing guide with troubleshooting |
| `server/cmd/test/main.go` | Test script to verify backend functionality |
| `server/cmd/cluelyd/main.go` | Main server entry point |
| `server/internal/answer/service.go` | Enhanced Gemini prompt implementation |
| `server/internal/ws/session.go` | WebSocket session with OCR context handling |

---

## 🔧 Advanced Configuration

### Optional Environment Variables

See `server/.env.example` for full list. Key optional vars:

```bash
# Use a different Gemini model
GEMINI_MODEL=gemini-1.5-pro

# Adjust ASR buffer size (if you see AUDIO_BACKPRESSURE warnings)
ASR_PCM_BUFFER=256

# Change server port
PORT=3000

# Adjust metrics logging frequency
METRICS_INTERVAL=60

# Set log level
LOG_LEVEL=debug
```

---

## 🐛 Troubleshooting

### "Failed to connect"
- Make sure the server is running
- Check port 8080 isn't blocked
- Try: `netstat -an | findstr :8080`

### "gemini status 401" or "403"
- Your API key is invalid or expired
- Get a new key from https://aistudio.google.com/app/apikey
- Verify you copied the entire key

### Generic fallback hints appear
- `GEMINI_API_KEY` not set correctly
- Check server logs for error messages
- Try setting as system environment variable

### No responses from Gemini
- Check rate limiting (1 hint per 1.5s)
- Look for `[answer]` errors in server logs
- Verify internet connection

For more troubleshooting, see **TESTING.md**.

---

## 🎓 Next Steps

Once your backend tests pass:

1. **Set up Google Cloud Speech-to-Text** for real ASR
   - Get credentials from: https://console.cloud.google.com/iam-admin/serviceaccounts
   - Set `GOOGLE_APPLICATION_CREDENTIALS` and `GCP_PROJECT_ID`

2. **Build the visionOS client**
   - Navigate to `apps/visionos`
   - Open in Xcode and build

3. **Connect client to server**
   - Update `Config.swift` with your server URL
   - Test with real conversations

4. **Monitor performance**
   - Watch metrics logs
   - Check for backpressure warnings
   - Tune buffer sizes if needed

---

## 📚 Additional Resources

- **TESTING.md** - Detailed testing procedures
- **PROJECT_STRUCTURE.md** - Codebase architecture overview
- **QUICKSTART.md** - Quick reference guide
- **implementation_status.md** - Current implementation status
- **general_guide.md** - General development guidelines

---

## 🙋 Need Help?

If you encounter issues:
1. Check the troubleshooting section in **TESTING.md**
2. Review server logs for error messages
3. Verify all environment variables are set correctly
4. Check that ports aren't blocked by firewall

---

**🎉 Congratulations! You now have a powerful AI-driven executive coaching backend!**

The enhanced Gemini prompt provides strategic, context-aware coaching hints that reference actual screen content and detect subtle communication signals. This makes Cluely a truly intelligent real-time coach for your conversations.

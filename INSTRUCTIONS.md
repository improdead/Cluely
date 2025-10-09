# Cluely — Instructions

**Cluely** is an AI-powered executive coaching assistant for Apple Vision Pro that provides real-time, context-aware coaching hints during conversations and presentations.

## 🎯 What is Cluely?

Cluely combines:
- **Go Backend**: WebSocket server with Google Cloud Speech-to-Text v2 + Gemini AI for intelligent micro-answers
- **visionOS Frontend**: Beautiful "liquid glass" UI with voice activity detection (VAD), audio streaming, and optional screen capture with OCR

The system listens to your conversations, analyzes what's being said (and optionally what's on screen), and provides strategic coaching hints in real-time.

## 📋 Quick Start

### Prerequisites

**Backend Requirements (Go Server):**
- Go 1.21+ installed
- **Gemini API Key** (required) - Get from [Google AI Studio](https://aistudio.google.com/app/apikey)
- **Google Cloud Credentials** (optional for Speech-to-Text):
  - GCP project with Speech-to-Text API enabled
  - Service account JSON with appropriate permissions
  - Environment variables: `GOOGLE_APPLICATION_CREDENTIALS`, `GCP_PROJECT_ID`

**Frontend Requirements (visionOS Client):**
- macOS with Xcode 15+
- visionOS SDK installed
- Vision Pro device or visionOS Simulator

### Installation & Setup

#### 1. Set Up Backend

```bash
cd server

# Set your Gemini API key (required)
export GEMINI_API_KEY="your-api-key-here"

# Optional: Set Google Cloud credentials for Speech-to-Text
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
export GCP_PROJECT_ID="your-gcp-project-id"

# Build the server
go build -o bin/cluelyd ./cmd/cluelyd

# Run the server
./bin/cluelyd
```

The server will start on port 8080. You should see:
```
cluelyd listening :8080
[metrics] sessions=0 pcm(in=0 drop=0) asr(p=0 f=0) hints=0 followups=0 errors(asr=0 ans=0)
```

#### 2. Set Up Frontend

On macOS:

1. Open Xcode and create a new visionOS App project named "Cluely"
2. Drag the contents of `apps/visionos/CluelyApp/` into your Xcode project
3. Update `Info.plist` with your server URL:
   - Same Mac: `ws://localhost:8080/ws`
   - Different machine: `ws://<SERVER_IP>:8080/ws`
4. Build and run in visionOS Simulator or device
5. Grant microphone permission when prompted

## 🧪 Testing

### Test Backend Only (No Frontend Required)

The easiest way to verify your backend works:

```bash
cd server

# Build and run the test client
go build -o bin/test ./cmd/test
./bin/test
```

This simulates a visionOS client and sends test scenarios with transcripts + OCR context. You should see intelligent, context-aware coaching hints.

**See [TESTING.md](TESTING.md) for detailed testing instructions and troubleshooting.**

### Test Full End-to-End

With both backend and frontend running:

**Option A: Voice Activity Detection (Auto)**
- Speak near the microphone
- VAD will detect speech and show "Listening..." indicator
- Real-time transcription → coaching hints appear
- Stops automatically after silence

**Option B: Manual Control**
- Tap "Start" to begin streaming
- Speak → see transcripts and hints
- Tap "Stop" to end

**Option C: Demo Mode**
- Tap "Demo" button to send test transcript
- Verify hint appears in suggestion panel

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the `server/` directory or set environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `GEMINI_API_KEY` | Yes | Your Gemini API key from Google AI Studio |
| `GEMINI_MODEL` | No | Model to use (default: `gemini-flash-lite-latest`) |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | Path to GCP service account JSON |
| `GCP_PROJECT_ID` | No | Your Google Cloud project ID |

See `server/.env.example` for a complete template.

### Client Configuration

Update `apps/visionos/CluelyApp/Resources/Info.plist`:
- `WSURL`: WebSocket server address (e.g., `ws://localhost:8080/ws`)

## 🎨 Features

### Current Features
- ✅ Beautiful liquid glass UI matching Vision Pro design language
- ✅ VAD-driven auto start/stop for hands-free operation
- ✅ Real-time Google Cloud Speech-to-Text streaming
- ✅ Gemini AI-powered strategic coaching hints
- ✅ ReplayKit + Vision OCR for screen context (optional)
- ✅ Rate limiting and error handling
- ✅ Manual controls and demo mode

### Planned Features
- 🔄 TLS/wss:// for production deployment
- 🔄 Session persistence (Redis/Postgres)
- 🔄 Advanced observability (OTel, pprof)
- 🔄 Auto-fade hint animations
- 🔄 Presenter-safe mode

## 🐛 Troubleshooting

### Backend Issues

**"gemini status 401" or "403"**
- Your API key is invalid or expired
- Get a new key from https://aistudio.google.com/app/apikey
- Verify you copied the entire key (starts with `AI...`)

**Generic fallback hints appear**
- `GEMINI_API_KEY` not set correctly
- Check server logs for error messages

**No ASR transcripts**
- Verify Google Cloud credentials are set
- Ensure Speech-to-Text API is enabled in GCP
- Check server logs for ASR errors

### Frontend Issues

**Client can't connect**
- Ensure server is running on expected host/port
- Check firewall rules (may block port 8080)
- Verify `Info.plist` WSURL matches server address

**VAD not triggering**
- Adjust thresholds in `VAD.swift` (`openDB`, `closeDB`)
- Verify microphone permission in Settings → Privacy & Security

**No hints appearing**
- Check server logs for Gemini errors
- Verify rate limiting (1 hint per 1.5s)
- Ensure GEMINI_API_KEY is set on server

### Port Already in Use

```bash
# Find process using port 8080
# macOS/Linux:
lsof -i :8080

# Windows:
netstat -ano | findstr :8080

# Kill the process or use a different port
```

## 📚 Documentation

This repository includes comprehensive documentation:

| File | Purpose |
|------|---------|
| **INSTRUCTIONS.md** | This file - quick start and overview |
| [QUICKSTART.md](QUICKSTART.md) | Detailed E2E setup guide |
| [SETUP_GUIDE.md](SETUP_GUIDE.md) | Backend setup and enhanced features |
| [TESTING.md](TESTING.md) | Comprehensive testing guide |
| [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) | Repository structure and file manifest |
| [general_guide.md](general_guide.md) | Complete system design and architecture |
| [implementation_status.md](implementation_status.md) | What's implemented vs. planned |
| [CHANGES.md](CHANGES.md) | Recent enhancements and changes |
| [server/README.md](server/README.md) | Server-specific documentation |

## 🚀 Next Steps

1. **Get Started**: Follow the Quick Start section above
2. **Test Backend**: Use the test script to verify Gemini integration
3. **Run Full Stack**: Set up both backend and frontend
4. **Explore Features**: Try different coaching scenarios
5. **Customize**: Adjust VAD thresholds, prompts, and UI to your needs

### Production Deployment

To deploy Cluely in production:
1. Set up TLS/wss:// using Caddy or a cloud proxy
2. Add Redis/Postgres for session persistence
3. Configure observability (OTel, pprof)
4. Deploy using Docker Compose (see `general_guide.md`)
5. Implement presenter-safe mode for screen capture

## 🙋 Need Help?

1. Check the troubleshooting sections in:
   - This file
   - [TESTING.md](TESTING.md)
   - [QUICKSTART.md](QUICKSTART.md)
2. Review server logs for detailed error messages
3. Verify all environment variables are set correctly
4. Ensure required services (GCP, Gemini) are accessible

## 🎉 What Makes Cluely Powerful

Cluely's enhanced Gemini prompt provides:

1. **Screen Content Awareness**: References actual OCR data from your presentations
2. **Subtle Signal Detection**: Catches hesitation words and low commitment signals
3. **Strategic Thinking**: Recognizes when audiences are confused and suggests pivots
4. **Tactical Next Steps**: Provides specific, actionable dialogue suggestions
5. **Simplification**: Uses analogies to make complex concepts accessible

**Example coaching hints:**
- *"'Probably' signals low commitment—confirm availability now"*
- *"Pivot to engagement growth story—higher LTV potential"*
- *"They're confused—simplify the explanation now"*
- *"Ask: 'Friday 2pm works—shall I send invite?'"*

Enjoy building with Cluely! 🎯

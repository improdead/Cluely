# Cluely visionOS App Setup

## Quick Setup Guide

### Required Frameworks

The Cluely app requires these Apple frameworks:

| Framework | Purpose |
|-----------|---------|
| **SwiftUI** | UI framework (included by default) |
| **AVFoundation** | Audio capture and processing |
| **Vision** | OCR text recognition from screen |
| **ReplayKit** | Screen capture functionality |
| **Accelerate** | Audio signal processing for VAD |

### File Structure

```
CluelyApp/
├── App.swift                    # Main app entry point
├── Config.swift                 # Configuration (WebSocket URL)
├── Input/
│   └── VAD.swift               # Voice Activity Detection
├── Session/
│   ├── AudioEngine.swift       # Audio capture pipeline
│   ├── CoachRuntime.swift      # Hint display logic
│   ├── FrameSource.swift       # Screen capture + OCR
│   ├── Models.swift            # Data models
│   ├── SessionController.swift # Main controller
│   └── WSClient.swift          # WebSocket client
├── Views/
│   ├── CoachChip.swift         # Bottom hint chip
│   ├── CoachView.swift         # Main coaching view
│   ├── ContextCard.swift       # Left context card
│   ├── FramePreviewCard.swift  # Frame capture indicator
│   ├── ListeningIndicator.swift # "Listening..." pill
│   ├── OrnamentControls.swift  # Bottom controls
│   ├── RootWindow.swift        # Main window
│   ├── SuggestionPanel.swift   # Right suggestions panel
│   └── WarningBadge.swift      # Warning overlay
└── Resources/
    └── Info.plist              # App configuration
```

## Setup Instructions

### Option 0: Open the Swift Package (Xcode 15+)

1. Open Xcode and choose **File → Open...**
2. Navigate to this repository and select `apps/visionos/Package.swift`
3. Xcode will materialise the package with the `Cluely` executable target preconfigured with the required frameworks
4. Pick the **Apple Vision Pro** simulator and build/run as usual

> If you're using an older Xcode that can't open Swift packages directly as apps, use the manual setup below instead.

### Option 1: Manual Xcode Setup (Recommended)

1. **Create New Project**
   ```
   - Open Xcode
   - File → New → Project
   - visionOS → App
   - Name: "Cluely"
   - Save to: apps/visionos/
   ```

2. **Remove Default Files**
   - Delete `ContentView.swift`

3. **Add Frameworks**
   - Select project → Target "Cluely"
   - "Frameworks, Libraries, and Embedded Content"
   - Click "+" and add:
     - AVFoundation.framework
     - Vision.framework
     - ReplayKit.framework
     - Accelerate.framework

4. **Add Source Files**
   - Right-click "Cluely" folder
   - "Add Files to Cluely..."
   - Select entire `CluelyApp/` folder
   - ✓ "Copy items if needed"
   - ✓ "Create groups"
   - Click "Add"

5. **Build & Run**
   - Select "Apple Vision Pro" simulator
   - Press ▶️ or Cmd+R

### Option 2: Using Setup Script

Run the setup script to see detailed instructions:

```bash
cd apps/visionos
./SETUP_XCODE.sh
```

## Configuration

### WebSocket URL

Edit `Resources/Info.plist`:

```xml
<key>WSURL</key>
<string>ws://localhost:8080/ws</string>
```

Change `localhost` to your server's IP if running on a different machine.

### Permissions

Required Info.plist keys (already configured):

- `NSMicrophoneUsageDescription` - For audio capture
- `WSURL` - WebSocket server address

## Testing

### 1. Demo Mode (No Server Required)
- Launch app
- Click "Demo" button
- Simulated hint should appear

### 2. With Server
```bash
# Start server first
cd ../../server
./bin/cluelyd

# Then run app in Xcode
```

### 3. Full Flow
- Enable "Frames" toggle for screen OCR
- Speak near microphone
- VAD triggers automatically
- Hints appear in real-time

## Troubleshooting

### Build Errors

**"Cannot find type 'AVAudioEngine'"**
- Missing AVFoundation framework
- Add in project settings → Frameworks

**"Cannot find type 'VNRecognizeTextRequest'"**
- Missing Vision framework
- Add in project settings → Frameworks

**"Cannot find type 'RPScreenRecorder'"**
- Missing ReplayKit framework
- Add in project settings → Frameworks

### Runtime Issues

**No microphone access**
- Check Settings → Privacy → Microphone
- Verify `NSMicrophoneUsageDescription` in Info.plist

**Can't connect to server**
- Verify server is running: `./bin/cluelyd`
- Check WSURL in Info.plist
- Ensure firewall allows port 8080

**VAD not triggering**
- Increase microphone volume
- Adjust thresholds in `Input/VAD.swift`

## Development Notes

### Architecture

- **SwiftUI + Combine**: Reactive UI updates
- **ObservableObject**: State management
- **@MainActor**: Thread-safe UI updates
- **Async/Await**: Modern concurrency

### Key Components

- **SessionController**: Orchestrates entire app
- **AudioEngine**: Captures & processes audio
- **FrameSource**: Screen capture with OCR
- **WSClient**: WebSocket communication
- **VAD**: Voice activity detection

### Performance

- Audio: 16kHz mono, 20ms frames (640 bytes)
- Screen OCR: 2 FPS (configurable)
- WebSocket: Binary (PCM) + JSON (control)
- Rate limiting: 1 hint per 1.5 seconds

## Next Steps

1. ✅ Set up Xcode project
2. ✅ Build and run app
3. ✅ Test demo mode
4. ✅ Connect to server
5. ✅ Test full voice → hint flow
6. 🔄 Customize VAD thresholds
7. 🔄 Adjust UI styling
8. 🔄 Add custom features

## Resources

- [Main Documentation](../../INSTRUCTIONS.md)
- [Backend Setup](../../server/README.md)
- [Testing Guide](../../TESTING.md)
- [Architecture Guide](../../general_guide.md)

---

**Need help? Check the troubleshooting sections or run `./SETUP_XCODE.sh` for detailed instructions.**

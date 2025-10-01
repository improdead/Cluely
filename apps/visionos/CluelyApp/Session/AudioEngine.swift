import AVFAudio

final class AudioEngine {
  private let engine = AVAudioEngine()
  private var converter: AVAudioConverter?
  private let outFormat = AVAudioFormat(commonFormat: .pcmFormatInt16,
                                        sampleRate: 16_000, channels: 1, interleaved: true)!
  func prepare() throws {
    let sess = AVAudioSession.sharedInstance()
    try sess.setCategory(.record, mode: .spokenAudio, options: [.duckOthers])
    try sess.setActive(true)
    let inFmt = engine.inputNode.inputFormat(forBus: 0)
    converter = AVAudioConverter(from: inFmt, to: outFormat)
  }
  // Start tap and emit both PCM bytes and an energy dB value per ~20ms frame
  func start(onFrame: @escaping (Data, Float)->Void) throws {
    let node = engine.inputNode
    let inFmt = node.inputFormat(forBus: 0)
    node.installTap(onBus: 0, bufferSize: 1024, format: inFmt) { [weak self] buf, _ in
      guard let self, let conv = self.converter else { return }
      let out = AVAudioPCMBuffer(pcmFormat: self.outFormat,
                                 frameCapacity: AVAudioFrameCount(320))! // 20ms
      var err: NSError?
      let inputBlock: AVAudioConverterInputBlock = { _, outStatus in
        outStatus.pointee = .haveData; return buf
      }
      conv.convert(to: out, error: &err, withInputFrom: inputBlock)
      guard err == nil, let ch = out.int16ChannelData else { return }
      let frames = Int(out.frameLength)
      if frames == 0 { return }
      // Build Data view
      let bytes = UnsafeBufferPointer(start: ch[0], count: frames).withMemoryRebound(to: UInt8.self) {
        Data(buffer: $0)
      }
      // Compute simple RMS dB from int16 samples
      var sum: Float = 0
      let p = ch[0]
      for i in 0..<frames {
        let v = Float(p[i]) / 32768.0
        sum += v * v
      }
      let rms = sqrtf(sum / Float(frames))
      let db = 20 * log10f(max(rms, 1e-9))
      onFrame(bytes, db) // 320 samples * 2 bytes = 640 bytes
    }
    try engine.start()
  }
  func stop() { engine.inputNode.removeTap(onBus: 0); engine.stop() }
}

import Accelerate
import AVFAudio

struct VADConfig { var openDB: Float = -42; var closeDB: Float = -48
  var minOpenMs: Int = 300; var hangMs: Int = 700 }

final class SimpleVAD {
  private var openedAt: Date?; private var lastSpeechAt: Date?
  func levelDB(_ b: AVAudioPCMBuffer) -> Float {
    guard let p = b.floatChannelData?[0] else { return -120 }
    var mean: Float = 0; vDSP_meamgv(p, 1, &mean, vDSP_Length(b.frameLength))
    let rms = sqrtf(mean); return 20 * log10f(max(rms, 1e-9))
  }
  func onFrame(db: Float, cfg: VADConfig) -> (open: Bool, close: Bool) {
    let now = Date()
    if openedAt == nil, db > cfg.openDB { openedAt = now; lastSpeechAt = now; return (true,false) }
    if let _ = openedAt {
      if db > cfg.closeDB { lastSpeechAt = now }
      if let last = lastSpeechAt, now.timeIntervalSince(last)*1000 > Double(cfg.hangMs) {
        openedAt = nil; lastSpeechAt = nil; return (false,true)
      }
    }
    return (false,false)
  }
}

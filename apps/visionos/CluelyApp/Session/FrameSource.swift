import Foundation
import ReplayKit
import Vision
import CoreImage

final class FrameSource {
  private var recorder: RPScreenRecorder?
  private var lastCaptureTime = Date.distantPast
  private let captureInterval: TimeInterval = 0.5 // 2 fps
  var onTokens: (([String]) -> Void)?

  func start() {
    recorder = RPScreenRecorder.shared()
    guard let rec = recorder else { return }
    // Start capture (broadcasts to self)
    rec.startCapture { [weak self] sampleBuffer, bufferType, error in
      guard let self, error == nil, bufferType == .video else { return }
      self.processSampleBuffer(sampleBuffer)
    } completionHandler: { err in
      if let e = err {
        print("[FrameSource] start capture error: \(e)")
      }
    }
  }

  func stop() {
    recorder?.stopCapture { err in
      if let e = err {
        print("[FrameSource] stop capture error: \(e)")
      }
    }
    recorder = nil
  }

  private func processSampleBuffer(_ buffer: CMSampleBuffer) {
    let now = Date()
    guard now.timeIntervalSince(lastCaptureTime) >= captureInterval else { return }
    lastCaptureTime = now

    guard let pixelBuffer = CMSampleBufferGetImageBuffer(buffer) else { return }
    let ciImage = CIImage(cvPixelBuffer: pixelBuffer)

    // Run VNRecognizeTextRequest
    let request = VNRecognizeTextRequest { [weak self] req, err in
      guard let observations = req.results as? [VNRecognizedTextObservation] else { return }
      var tokens: [String] = []
      for obs in observations.prefix(10) { // top 10
        if let text = obs.topCandidates(1).first?.string {
          tokens.append(text)
        }
      }
      if !tokens.isEmpty {
        self?.onTokens?(tokens)
      }
    }
    request.recognitionLevel = .fast
    request.usesLanguageCorrection = false

    let handler = VNImageRequestHandler(ciImage: ciImage, options: [:])
    try? handler.perform([request])
  }
}

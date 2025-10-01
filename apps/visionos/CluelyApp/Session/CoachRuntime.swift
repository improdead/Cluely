import Foundation

final class CoachRuntime {
  private var lastShownAt = Date.distantPast
  private var lastValue = ""
  private let minInterval: TimeInterval = 1.5

  func shouldShow(_ t: String) -> Bool {
    let n = CoachRuntime.normalize(t)
    if n.isEmpty { return false }
    let now = Date()
    if n == lastValue { return false }
    if now.timeIntervalSince(lastShownAt) < minInterval { return false }
    lastShownAt = now
    lastValue = n
    return true
  }
  static func normalize(_ s: String) -> String {
    let lowered = s.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
    let filtered = lowered.filter { $0.isLetter || $0.isNumber || $0.isWhitespace }
    return filtered
  }
}

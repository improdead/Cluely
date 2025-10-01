import Foundation

struct ServerEvent: Codable {
  let type: String
  let listening: Bool?
  let text: String?
  let ttlMs: Int?
  let code: String?
  let msg: String?
}

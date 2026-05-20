import Foundation
import Vapor

struct Paste: Codable, Sendable {
    var id: String
    var title: String
    var content: String
    var language: String
    var createdAt: Date
    var expiresAt: Date?
    var passwordHash: String?
    var isProtected: Bool
}

struct PasteStoreKey: StorageKey {
    typealias Value = PasteStore
}

extension Application {
    var pasteStore: PasteStore {
        get {
            guard let store = storage[PasteStoreKey.self] else {
                fatalError("PasteStore not configured")
            }
            return store
        }
        set {
            storage[PasteStoreKey.self] = newValue
        }
    }
}

actor PasteStore {
    private var pastes: [String: Paste] = [:]

    func create(_ paste: inout Paste) {
        paste.id = Self.generateID()
        paste.createdAt = Date()
        pastes[paste.id] = paste
    }

    func get(_ id: String) -> Paste? {
        pastes[id]
    }

    func delete(_ id: String) {
        pastes.removeValue(forKey: id)
    }

    func startExpiryWorker() async {
        while true {
            try? await Task.sleep(nanoseconds: 30_000_000_000)
            let now = Date()
            let expired = pastes.compactMap { id, p -> String? in
                guard let exp = p.expiresAt, now > exp else { return nil }
                return id
            }
            for id in expired {
                pastes.removeValue(forKey: id)
            }
        }
    }

    private static func generateID() -> String {
        let chars = Array("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
        return String((0..<8).map { _ in chars.randomElement()! })
    }
}

import Foundation
import Vapor

struct Paste: Codable, Sendable {
    var id: String
    let title: String
    let content: String
    let language: String
    var createdAt: Date
    let expiresAt: Date?
    let passwordHash: String?
    let isProtected: Bool
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

    @discardableResult
    func create(_ paste: Paste) -> String {
        var stored = paste
        stored.id = Self.generateID()
        stored.createdAt = Date()
        pastes[stored.id] = stored
        return stored.id
    }

    func get(_ id: String) -> Paste? {
        pastes[id]
    }

    func delete(_ id: String) {
        pastes.removeValue(forKey: id)
    }

    func startExpiryWorker() async {
        do {
            while true {
                try await Task.sleep(nanoseconds: 30_000_000_000)
                let now = Date()
                let expired = pastes.compactMap { id, paste -> String? in
                    guard let exp = paste.expiresAt, now > exp else { return nil }
                    return id
                }
                for id in expired {
                    pastes.removeValue(forKey: id)
                }
            }
        } catch {
            // task was cancelled on shutdown
        }
    }

    private static func generateID() -> String {
        let chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
        return String((0..<8).compactMap { _ in chars.randomElement() })
    }
}

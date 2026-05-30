import Vapor

final class ExpiryLifecycle: LifecycleHandler {
    private let store: PasteStore
    nonisolated(unsafe) private var task: Task<Void, Never>?

    init(store: PasteStore) {
        self.store = store
    }

    func didBoot(_ application: Application) throws {
        task = Task { await store.startExpiryWorker() }
    }

    func shutdown(_ application: Application) {
        task?.cancel()
    }
}

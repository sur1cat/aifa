import SwiftUI

@main
struct AtomaWatchApp: App {
    @StateObject private var dataStore = WatchDataStore.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(dataStore)
        }
    }
}

import SwiftUI

struct ContentView: View {
    @EnvironmentObject var dataStore: WatchDataStore

    var body: some View {
        TabView {
            HabitsWatchView()
            TasksWatchView()
            BudgetWatchView()
        }
        .tabViewStyle(.verticalPage)
    }
}

#Preview {
    ContentView()
        .environmentObject(WatchDataStore.shared)
}

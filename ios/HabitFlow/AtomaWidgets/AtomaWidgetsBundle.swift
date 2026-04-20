import WidgetKit
import SwiftUI

@main
struct AtomaWidgetsBundle: WidgetBundle {
    var body: some Widget {
        HabitsWidget()
        TasksWidget()
        BudgetWidget()
    }
}

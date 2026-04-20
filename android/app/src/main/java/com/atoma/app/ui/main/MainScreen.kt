package com.atoma.app.ui.main

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.ui.unit.dp
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.atoma.app.R
import com.atoma.app.data.repository.AIAgent
import com.atoma.app.ui.budget.BudgetScreen
import com.atoma.app.ui.components.OfflineBanner
import com.atoma.app.ui.components.SyncErrorBanner
import com.atoma.app.ui.components.SyncingBanner
import com.atoma.app.ui.dashboard.DashboardScreen
import com.atoma.app.ui.habits.HabitsScreen
import com.atoma.app.ui.profile.ProfileScreen
import com.atoma.app.ui.tasks.TasksScreen
import com.atoma.app.ui.theme.Primary

sealed class BottomNavItem(
    val route: String,
    val titleResId: Int,
    val selectedIcon: ImageVector,
    val unselectedIcon: ImageVector
) {
    data object Dashboard : BottomNavItem(
        route = "dashboard",
        titleResId = R.string.nav_dashboard,
        selectedIcon = Icons.Filled.Dashboard,
        unselectedIcon = Icons.Outlined.Dashboard
    )

    data object Habits : BottomNavItem(
        route = "habits",
        titleResId = R.string.nav_habits,
        selectedIcon = Icons.Filled.Loop,
        unselectedIcon = Icons.Outlined.Loop
    )

    data object Tasks : BottomNavItem(
        route = "tasks",
        titleResId = R.string.nav_tasks,
        selectedIcon = Icons.Filled.CheckCircle,
        unselectedIcon = Icons.Outlined.CheckCircle
    )

    data object Budget : BottomNavItem(
        route = "budget",
        titleResId = R.string.nav_budget,
        selectedIcon = Icons.Filled.AccountBalanceWallet,
        unselectedIcon = Icons.Outlined.AccountBalanceWallet
    )

    data object Profile : BottomNavItem(
        route = "profile",
        titleResId = R.string.nav_profile,
        selectedIcon = Icons.Filled.Person,
        unselectedIcon = Icons.Outlined.Person
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainScreen(
    onLogout: () -> Unit,
    onNavigateToAI: (AIAgent) -> Unit,
    viewModel: MainViewModel = hiltViewModel()
) {
    val navController = rememberNavController()
    val navItems = listOf(
        BottomNavItem.Dashboard,
        BottomNavItem.Habits,
        BottomNavItem.Tasks,
        BottomNavItem.Budget,
        BottomNavItem.Profile
    )

    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route
    val syncState by viewModel.syncState.collectAsState()

    // Determine which AI agent to use based on current tab
    val currentAgent = when (currentRoute) {
        BottomNavItem.Dashboard.route -> AIAgent.LIFE_COACH
        BottomNavItem.Habits.route -> AIAgent.HABIT_COACH
        BottomNavItem.Tasks.route -> AIAgent.TASK_ASSISTANT
        BottomNavItem.Budget.route -> AIAgent.FINANCE_ADVISOR
        else -> AIAgent.LIFE_COACH
    }

    Scaffold(
        topBar = {
            Column {
                OfflineBanner(isVisible = !syncState.isOnline)
                SyncingBanner(isVisible = syncState.isSyncing)
                SyncErrorBanner(error = syncState.error)
            }
        },
        floatingActionButton = {
            if (currentRoute != BottomNavItem.Profile.route) {
                FloatingActionButton(
                    onClick = { onNavigateToAI(currentAgent) },
                    containerColor = Primary,
                    contentColor = Color.White
                ) {
                    Icon(
                        imageVector = Icons.Default.AutoAwesome,
                        contentDescription = "AI Assistant"
                    )
                }
            }
        },
        bottomBar = {
            NavigationBar(
                containerColor = MaterialTheme.colorScheme.surface,
                tonalElevation = 0.dp
            ) {
                val currentDestination = navBackStackEntry?.destination

                navItems.forEach { item ->
                    val selected = currentDestination?.hierarchy?.any { it.route == item.route } == true

                    NavigationBarItem(
                        icon = {
                            Icon(
                                imageVector = if (selected) item.selectedIcon else item.unselectedIcon,
                                contentDescription = stringResource(item.titleResId)
                            )
                        },
                        label = {
                            Text(
                                text = stringResource(item.titleResId),
                                style = MaterialTheme.typography.labelMedium
                            )
                        },
                        selected = selected,
                        onClick = {
                            navController.navigate(item.route) {
                                popUpTo(navController.graph.findStartDestination().id) {
                                    saveState = true
                                }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        colors = NavigationBarItemDefaults.colors(
                            selectedIconColor = MaterialTheme.colorScheme.primary,
                            selectedTextColor = MaterialTheme.colorScheme.primary,
                            indicatorColor = MaterialTheme.colorScheme.primaryContainer
                        )
                    )
                }
            }
        }
    ) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = BottomNavItem.Dashboard.route,
            modifier = Modifier.padding(innerPadding)
        ) {
            composable(BottomNavItem.Dashboard.route) {
                DashboardScreen()
            }
            composable(BottomNavItem.Habits.route) {
                HabitsScreen()
            }
            composable(BottomNavItem.Tasks.route) {
                TasksScreen()
            }
            composable(BottomNavItem.Budget.route) {
                BudgetScreen()
            }
            composable(BottomNavItem.Profile.route) {
                ProfileScreen(onLogout = onLogout)
            }
        }
    }
}

package com.atoma.app.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.atoma.app.data.repository.AIAgent
import com.atoma.app.ui.ai.AIChatScreen
import com.atoma.app.ui.auth.AuthViewModel
import com.atoma.app.ui.auth.LoginScreen
import com.atoma.app.ui.main.MainScreen
import com.atoma.app.ui.onboarding.OnboardingScreen
import com.atoma.app.ui.onboarding.OnboardingViewModel

sealed class Screen(val route: String) {
    data object Onboarding : Screen("onboarding")
    data object Login : Screen("login")
    data object Main : Screen("main")
    data object AIChat : Screen("ai_chat/{agent}") {
        fun createRoute(agent: AIAgent) = "ai_chat/${agent.id}"
    }
}

@Composable
fun AtomaNavHost() {
    val navController = rememberNavController()
    val authViewModel: AuthViewModel = hiltViewModel()
    val onboardingViewModel: OnboardingViewModel = hiltViewModel()
    val isLoggedIn by authViewModel.isLoggedIn.collectAsState()
    val hasCompletedOnboarding by onboardingViewModel.hasCompletedOnboarding.collectAsState()

    val startDestination = when {
        !hasCompletedOnboarding -> Screen.Onboarding.route
        isLoggedIn -> Screen.Main.route
        else -> Screen.Login.route
    }

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        composable(Screen.Onboarding.route) {
            OnboardingScreen(
                onComplete = {
                    navController.navigate(Screen.Login.route) {
                        popUpTo(Screen.Onboarding.route) { inclusive = true }
                    }
                }
            )
        }

        composable(Screen.Login.route) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate(Screen.Main.route) {
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                }
            )
        }

        composable(Screen.Main.route) {
            MainScreen(
                onLogout = {
                    navController.navigate(Screen.Login.route) {
                        popUpTo(Screen.Main.route) { inclusive = true }
                    }
                },
                onNavigateToAI = { agent ->
                    navController.navigate(Screen.AIChat.createRoute(agent))
                }
            )
        }

        composable(
            route = Screen.AIChat.route,
            arguments = listOf(navArgument("agent") { type = NavType.StringType })
        ) { backStackEntry ->
            val agentId = backStackEntry.arguments?.getString("agent") ?: "life_coach"
            val agent = AIAgent.entries.find { it.id == agentId } ?: AIAgent.LIFE_COACH
            AIChatScreen(
                initialAgent = agent,
                onBack = { navController.popBackStack() }
            )
        }
    }
}

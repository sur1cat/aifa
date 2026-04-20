package com.atoma.app.ui.dashboard

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import com.atoma.app.domain.model.DailyTask
import com.atoma.app.domain.model.Habit
import com.atoma.app.domain.model.TaskPriority
import com.atoma.app.ui.components.CollapsibleSectionHeader
import com.atoma.app.ui.theme.CornerRadius
import com.atoma.app.ui.theme.Primary
import com.atoma.app.ui.theme.ProgressSize
import com.atoma.app.ui.theme.Spacing
import java.time.format.DateTimeFormatter

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DashboardScreen(
    viewModel: DashboardViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    var upNextExpanded by remember { mutableStateOf(true) }
    var streaksExpanded by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) {
        viewModel.loadData()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(
                            text = "Atoma",
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            text = "your daily overview",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(Spacing.md),
            verticalArrangement = Arrangement.spacedBy(Spacing.md)
        ) {
            // Today's Progress Card
            item {
                TodayProgressCard(
                    greeting = uiState.greeting,
                    statusMessage = uiState.statusMessage,
                    habitsCompleted = uiState.habitsCompletedToday,
                    habitsTotal = uiState.habitsTotalToday,
                    tasksCompleted = uiState.tasksCompletedToday,
                    tasksTotal = uiState.tasksTotalToday,
                    progress = uiState.totalProgress.toFloat()
                )
            }

            // Life Score Card
            item {
                LifeScoreCard(
                    score = uiState.lifeScore,
                    habitsScore = uiState.habitsScore,
                    tasksScore = uiState.tasksScore,
                    budgetScore = uiState.budgetScore
                )
            }

            // Quick Stats Row
            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(Spacing.sm)
                ) {
                    QuickStatCard(
                        modifier = Modifier.weight(1f),
                        title = "Habits",
                        value = "${uiState.habitsCompletedToday}/${uiState.habitsTotalToday}",
                        subtitle = if (uiState.habitsTotalToday > 0) "${(uiState.habitsProgress * 100).toInt()}%" else "No habits",
                        icon = Icons.Default.Loop,
                        color = Primary
                    )
                    QuickStatCard(
                        modifier = Modifier.weight(1f),
                        title = "Tasks",
                        value = "${uiState.tasksCompletedToday}/${uiState.tasksTotalToday}",
                        subtitle = if (uiState.tasksTotalToday > 0) "${(uiState.tasksProgress * 100).toInt()}%" else "No tasks",
                        icon = Icons.Default.CheckCircle,
                        color = MaterialTheme.colorScheme.tertiary
                    )
                }
            }

            // Pending Items with Collapsible Header
            if (uiState.pendingHabits.isNotEmpty() || uiState.pendingTasks.isNotEmpty()) {
                item {
                    CollapsibleSectionHeader(
                        title = "Up Next",
                        expanded = upNextExpanded,
                        onToggle = { upNextExpanded = !upNextExpanded },
                        icon = Icons.Default.PlayArrow,
                        count = uiState.pendingHabits.size + uiState.pendingTasks.size
                    )
                }

                item {
                    AnimatedVisibility(
                        visible = upNextExpanded,
                        enter = expandVertically() + fadeIn(),
                        exit = shrinkVertically() + fadeOut()
                    ) {
                        Column(verticalArrangement = Arrangement.spacedBy(Spacing.xs)) {
                            uiState.pendingHabits.take(3).forEach { habit ->
                                PendingHabitRow(habit = habit)
                            }
                            uiState.pendingTasks.take(3).forEach { task ->
                                PendingTaskRow(task = task)
                            }
                        }
                    }
                }
            }

            // Monthly Budget Card
            item {
                MonthBudgetCard(
                    income = uiState.monthIncome,
                    expenses = uiState.monthExpenses,
                    balance = uiState.monthBalance
                )
            }

            // Active Streaks with Collapsible Header
            if (uiState.topStreakHabits.isNotEmpty()) {
                item {
                    CollapsibleSectionHeader(
                        title = "Active Streaks",
                        expanded = streaksExpanded,
                        onToggle = { streaksExpanded = !streaksExpanded },
                        icon = Icons.Default.LocalFireDepartment,
                        iconTint = Color(0xFFFF9800),
                        count = uiState.topStreakHabits.size
                    )
                }

                item {
                    AnimatedVisibility(
                        visible = streaksExpanded,
                        enter = expandVertically() + fadeIn(),
                        exit = shrinkVertically() + fadeOut()
                    ) {
                        Column(verticalArrangement = Arrangement.spacedBy(Spacing.xs)) {
                            uiState.topStreakHabits.take(3).forEach { habit ->
                                StreakRow(habit = habit)
                            }
                        }
                    }
                }
            }

            // Bottom spacing
            item {
                Spacer(modifier = Modifier.height(80.dp))
            }
        }
    }
}

@Composable
fun TodayProgressCard(
    greeting: String,
    statusMessage: String,
    habitsCompleted: Int,
    habitsTotal: Int,
    tasksCompleted: Int,
    tasksTotal: Int,
    progress: Float
) {
    val animatedProgress by animateFloatAsState(
        targetValue = progress,
        animationSpec = tween(durationMillis = 500),
        label = "progress"
    )

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Text(
                    text = greeting,
                    style = MaterialTheme.typography.headlineSmall,
                    fontWeight = FontWeight.Bold
                )
                Text(
                    text = statusMessage,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            if (habitsTotal + tasksTotal > 0) {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    LinearProgressIndicator(
                        progress = { animatedProgress },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(8.dp)
                            .clip(RoundedCornerShape(4.dp)),
                        color = Primary,
                        trackColor = Primary.copy(alpha = 0.15f),
                        strokeCap = StrokeCap.Round
                    )

                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(
                            text = "${habitsCompleted + tasksCompleted} of ${habitsTotal + tasksTotal} completed",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Text(
                            text = "${(animatedProgress * 100).toInt()}%",
                            style = MaterialTheme.typography.bodySmall,
                            fontWeight = FontWeight.SemiBold,
                            color = Primary
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun LifeScoreCard(
    score: Double,
    habitsScore: Double,
    tasksScore: Double,
    budgetScore: Double
) {
    val animatedScore by animateFloatAsState(
        targetValue = score.toFloat(),
        animationSpec = tween(durationMillis = 1000),
        label = "score"
    )

    val scoreColor = when {
        animatedScore >= 80 -> Primary
        animatedScore >= 60 -> MaterialTheme.colorScheme.tertiary
        animatedScore >= 40 -> Color(0xFFFFA000)
        else -> MaterialTheme.colorScheme.error
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Score Ring
            Box(
                modifier = Modifier.size(120.dp),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator(
                    progress = { 1f },
                    modifier = Modifier.fillMaxSize(),
                    color = scoreColor.copy(alpha = 0.2f),
                    strokeWidth = 12.dp,
                    strokeCap = StrokeCap.Round
                )
                CircularProgressIndicator(
                    progress = { animatedScore / 100f },
                    modifier = Modifier.fillMaxSize(),
                    color = scoreColor,
                    strokeWidth = 12.dp,
                    strokeCap = StrokeCap.Round
                )
                Text(
                    text = "${animatedScore.toInt()}",
                    style = MaterialTheme.typography.headlineLarge,
                    fontWeight = FontWeight.Bold,
                    color = scoreColor
                )
            }

            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                Text(
                    text = "Life Score",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Medium
                )
                Text(
                    text = "Your weekly balance",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // Breakdown
            Row(
                horizontalArrangement = Arrangement.spacedBy(24.dp)
            ) {
                BreakdownItem(
                    icon = Icons.Default.Loop,
                    value = habitsScore.toInt(),
                    color = Primary
                )
                BreakdownItem(
                    icon = Icons.Default.CheckCircle,
                    value = tasksScore.toInt(),
                    color = MaterialTheme.colorScheme.tertiary
                )
                BreakdownItem(
                    icon = Icons.Default.AccountBalanceWallet,
                    value = budgetScore.toInt(),
                    color = Color(0xFF4CAF50)
                )
            }
        }
    }
}

@Composable
fun BreakdownItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    value: Int,
    color: Color
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(4.dp)
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            modifier = Modifier.size(14.dp),
            tint = color
        )
        Text(
            text = "$value%",
            style = MaterialTheme.typography.bodySmall,
            fontWeight = FontWeight.Medium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun QuickStatCard(
    modifier: Modifier = Modifier,
    title: String,
    value: String,
    subtitle: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    color: Color
) {
    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(CornerRadius.medium)
    ) {
        Column(
            modifier = Modifier.padding(Spacing.sm),
            verticalArrangement = Arrangement.spacedBy(Spacing.xs)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = icon,
                    contentDescription = null,
                    modifier = Modifier.size(16.dp),
                    tint = color
                )
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Medium,
                    color = color
                )
            }
            Text(
                text = value,
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )
            Text(
                text = title,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
fun PendingHabitRow(habit: Habit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 14.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = habit.icon,
                fontSize = 18.sp
            )
            Text(
                text = habit.title,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.weight(1f)
            )
            Surface(
                color = Primary.copy(alpha = 0.1f),
                shape = RoundedCornerShape(50)
            ) {
                Text(
                    text = "Habit",
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Medium,
                    color = Primary,
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                )
            }
        }
    }
}

@Composable
fun PendingTaskRow(task: DailyTask) {
    val priorityColor = when (task.priority) {
        TaskPriority.URGENT -> Color(0xFFDC2626)
        TaskPriority.HIGH -> MaterialTheme.colorScheme.error
        TaskPriority.MEDIUM -> Color(0xFFFFA000)
        TaskPriority.LOW -> MaterialTheme.colorScheme.tertiary
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 14.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Icon(
                imageVector = Icons.Default.Circle,
                contentDescription = null,
                modifier = Modifier.size(18.dp),
                tint = priorityColor.copy(alpha = 0.7f)
            )
            Text(
                text = task.title,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.weight(1f)
            )
            Surface(
                color = priorityColor.copy(alpha = 0.1f),
                shape = RoundedCornerShape(50)
            ) {
                Text(
                    text = task.priority.title,
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Medium,
                    color = priorityColor,
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                )
            }
        }
    }
}

@Composable
fun MonthBudgetCard(
    income: Double,
    expenses: Double,
    balance: Double
) {
    val monthName = remember {
        java.time.LocalDate.now().format(DateTimeFormatter.ofPattern("MMMM"))
    }

    fun formatAmount(amount: Double): String {
        val absAmount = kotlin.math.abs(amount)
        return when {
            absAmount >= 1_000_000 -> "$${String.format("%.1fM", absAmount / 1_000_000)}"
            absAmount >= 1000 -> "$${String.format("%.1fK", absAmount / 1000)}"
            else -> "$${absAmount.toInt()}"
        }
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(14.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "$monthName Budget",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    text = if (balance >= 0) "+${formatAmount(balance)}" else "-${formatAmount(balance)}",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
                    color = if (balance >= 0) Color(0xFF4CAF50) else MaterialTheme.colorScheme.error
                )
            }

            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .background(Color(0xFF4CAF50), CircleShape)
                    )
                    Text(
                        text = "Income: ${formatAmount(income)}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .background(MaterialTheme.colorScheme.error, CircleShape)
                    )
                    Text(
                        text = "Expenses: ${formatAmount(expenses)}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

@Composable
fun StreakRow(habit: Habit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 14.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = habit.icon,
                fontSize = 18.sp
            )
            Text(
                text = habit.title,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.weight(1f)
            )
            Surface(
                color = Color(0xFFFF9800).copy(alpha = 0.1f),
                shape = RoundedCornerShape(50)
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.LocalFireDepartment,
                        contentDescription = "Streak",
                        modifier = Modifier.size(12.dp),
                        tint = Color(0xFFFF9800)
                    )
                    Text(
                        text = "${habit.streak}",
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

package com.atoma.app.ui.habits

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.LocalFireDepartment
import androidx.compose.material.icons.filled.Unarchive
import androidx.compose.material.icons.outlined.Archive
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import com.atoma.app.R
import com.atoma.app.domain.model.Goal
import com.atoma.app.domain.model.Habit
import com.atoma.app.ui.components.AnimatedCheckmark
import com.atoma.app.ui.goals.GoalsViewModel
import com.atoma.app.ui.theme.CornerRadius
import com.atoma.app.ui.theme.Primary
import com.atoma.app.ui.theme.Spacing
import java.time.LocalDate

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HabitsScreen(
    habitsViewModel: HabitsViewModel = hiltViewModel(),
    goalsViewModel: GoalsViewModel = hiltViewModel()
) {
    val habitsState by habitsViewModel.uiState.collectAsState()
    val goalsState by goalsViewModel.uiState.collectAsState()

    var showAddGoalDialog by remember { mutableStateOf(false) }
    var selectedDate by remember { mutableStateOf(LocalDate.now()) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(
                            text = stringResource(R.string.habits_title),
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            text = stringResource(R.string.habits_subtitle),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                },
                actions = {
                    if (habitsState.archivedHabits.isNotEmpty() || goalsState.archivedGoals.isNotEmpty()) {
                        IconButton(onClick = {
                            if (habitsState.archivedHabits.isNotEmpty()) {
                                habitsViewModel.showArchivedSheet()
                            } else {
                                goalsViewModel.showArchivedSheet()
                            }
                        }) {
                            Icon(Icons.Outlined.Archive, contentDescription = "Archived")
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { habitsViewModel.showAddDialog() },
                containerColor = MaterialTheme.colorScheme.primary
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.add_habit))
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Date Selector
            HabitsDateSelector(
                selectedDate = selectedDate,
                onDateSelected = { selectedDate = it }
            )

            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(bottom = 80.dp)
            ) {
                // Goals Section
                item {
                    GoalsSection(
                        goals = goalsState.goals,
                        onAddGoal = { showAddGoalDialog = true },
                        onArchiveGoal = { goalsViewModel.archiveGoal(it) }
                    )
                }

                // Habits Section Header
                item {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = Spacing.md, vertical = Spacing.xs),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = "Habits",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold
                        )
                        Text(
                            text = "${habitsState.habits.count { it.isCompletedToday }}/${habitsState.habits.size} today",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }

                when {
                    habitsState.isLoading && habitsState.habits.isEmpty() -> {
                        item {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(200.dp),
                                contentAlignment = Alignment.Center
                            ) {
                                CircularProgressIndicator()
                            }
                        }
                    }
                    habitsState.habits.isEmpty() -> {
                        item {
                            EmptyHabitsView(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(Spacing.xxl)
                            )
                        }
                    }
                    else -> {
                        items(habitsState.habits, key = { it.id }) { habit ->
                            HabitCard(
                                habit = habit,
                                selectedDate = selectedDate,
                                onToggle = { habitsViewModel.toggleHabit(habit, selectedDate) },
                                onClick = { /* TODO: show detail */ },
                                onArchive = { habitsViewModel.archiveHabit(habit) },
                                modifier = Modifier.padding(horizontal = Spacing.md, vertical = Spacing.xxs)
                            )
                        }
                    }
                }
            }
        }
    }

    if (habitsState.showAddDialog) {
        AddHabitDialog(
            onDismiss = { habitsViewModel.hideAddDialog() },
            onConfirm = { title, icon, color, period ->
                habitsViewModel.createHabit(title, icon, color, period)
            }
        )
    }

    if (showAddGoalDialog) {
        AddGoalDialog(
            onDismiss = { showAddGoalDialog = false },
            onConfirm = { title, icon ->
                goalsViewModel.createGoal(title, icon)
                showAddGoalDialog = false
            }
        )
    }

    if (goalsState.showArchivedSheet) {
        ArchivedGoalsSheet(
            goals = goalsState.archivedGoals,
            onDismiss = { goalsViewModel.hideArchivedSheet() },
            onUnarchive = { goalsViewModel.unarchiveGoal(it) },
            onDelete = { goalsViewModel.deleteGoal(it) }
        )
    }

    if (habitsState.showArchivedSheet) {
        ArchivedHabitsSheet(
            habits = habitsState.archivedHabits,
            onDismiss = { habitsViewModel.hideArchivedSheet() },
            onUnarchive = { habitsViewModel.unarchiveHabit(it) },
            onDelete = { habitsViewModel.deleteHabit(it) }
        )
    }
}

@Composable
private fun GoalsSection(
    goals: List<Goal>,
    onAddGoal: () -> Unit,
    onArchiveGoal: (Goal) -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "Goals",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )
            TextButton(onClick = onAddGoal) {
                Icon(
                    Icons.Default.Add,
                    contentDescription = null,
                    modifier = Modifier.size(18.dp)
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text("Add")
            }
        }

        if (goals.isEmpty()) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 8.dp),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
                )
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onAddGoal() }
                        .padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(text = "🎯", fontSize = 24.sp)
                    Spacer(modifier = Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = "Set your first goal",
                            style = MaterialTheme.typography.bodyMedium,
                            fontWeight = FontWeight.Medium
                        )
                        Text(
                            text = "Track progress with habits",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                    Icon(
                        Icons.Default.ChevronRight,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        } else {
            LazyRow(
                contentPadding = PaddingValues(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(goals, key = { it.id }) { goal ->
                    GoalCard(
                        goal = goal,
                        onArchive = { onArchiveGoal(goal) }
                    )
                }

                item {
                    AddGoalCard(onClick = onAddGoal)
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun GoalCard(
    goal: Goal,
    onArchive: () -> Unit
) {
    var showMenu by remember { mutableStateOf(false) }

    Card(
        onClick = { showMenu = true },
        modifier = Modifier.width(140.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = Primary.copy(alpha = 0.1f)
        )
    ) {
        Column(
            modifier = Modifier.padding(12.dp)
        ) {
            Text(text = goal.icon, fontSize = 28.sp)
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = goal.title,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis
            )
        }

        DropdownMenu(
            expanded = showMenu,
            onDismissRequest = { showMenu = false }
        ) {
            DropdownMenuItem(
                text = { Text("Archive") },
                onClick = {
                    showMenu = false
                    onArchive()
                },
                leadingIcon = {
                    Icon(Icons.Outlined.Archive, contentDescription = null)
                }
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AddGoalCard(onClick: () -> Unit) {
    Card(
        onClick = onClick,
        modifier = Modifier.width(100.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
        )
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                Icons.Default.Add,
                contentDescription = null,
                modifier = Modifier.size(28.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Add",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AddGoalDialog(
    onDismiss: () -> Unit,
    onConfirm: (String, String) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var icon by remember { mutableStateOf("🎯") }

    val icons = listOf("🎯", "💪", "📚", "💰", "🏃", "🧘", "✈️", "🎨", "💡", "⭐")

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("New Goal") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text("Goal title") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )

                Text(
                    text = "Icon",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    icons.take(5).forEach { emoji ->
                        Surface(
                            onClick = { icon = emoji },
                            shape = CircleShape,
                            color = if (icon == emoji) Primary.copy(alpha = 0.2f)
                            else MaterialTheme.colorScheme.surfaceVariant
                        ) {
                            Text(
                                text = emoji,
                                modifier = Modifier.padding(8.dp),
                                fontSize = 20.sp
                            )
                        }
                    }
                }

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    icons.drop(5).forEach { emoji ->
                        Surface(
                            onClick = { icon = emoji },
                            shape = CircleShape,
                            color = if (icon == emoji) Primary.copy(alpha = 0.2f)
                            else MaterialTheme.colorScheme.surfaceVariant
                        ) {
                            Text(
                                text = emoji,
                                modifier = Modifier.padding(8.dp),
                                fontSize = 20.sp
                            )
                        }
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = { onConfirm(title, icon) },
                enabled = title.isNotBlank(),
                colors = ButtonDefaults.buttonColors(containerColor = Primary)
            ) {
                Text("Create")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun ArchivedGoalsSheet(
    goals: List<Goal>,
    onDismiss: () -> Unit,
    onUnarchive: (Goal) -> Unit,
    onDelete: (Goal) -> Unit
) {
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Archived Goals",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(16.dp))

            goals.forEach { goal ->
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 4.dp),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(text = goal.icon, fontSize = 24.sp)

                        Spacer(modifier = Modifier.width(12.dp))

                        Text(
                            text = goal.title,
                            modifier = Modifier.weight(1f),
                            style = MaterialTheme.typography.bodyLarge
                        )

                        IconButton(onClick = { onUnarchive(goal) }) {
                            Icon(
                                Icons.Default.Unarchive,
                                contentDescription = "Restore",
                                tint = Primary
                            )
                        }

                        IconButton(onClick = { onDelete(goal) }) {
                            Icon(
                                Icons.Default.Delete,
                                contentDescription = "Delete",
                                tint = MaterialTheme.colorScheme.error
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HabitCard(
    habit: Habit,
    selectedDate: LocalDate,
    onToggle: () -> Unit,
    onClick: () -> Unit,
    onArchive: () -> Unit,
    modifier: Modifier = Modifier
) {
    var showMenu by remember { mutableStateOf(false) }
    val isCompletedOnDate = habit.completedDates.contains(selectedDate.toString())

    Card(
        modifier = modifier
            .fillMaxWidth()
            .clickable { showMenu = true },
        shape = RoundedCornerShape(CornerRadius.large),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(Spacing.md),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Icon
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(RoundedCornerShape(CornerRadius.medium))
                    .background(habit.habitColor.copy(alpha = 0.15f)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = habit.icon.toEmojiOrDefault(),
                    style = MaterialTheme.typography.titleLarge
                )
            }

            Spacer(modifier = Modifier.width(Spacing.md))

            // Title & Info
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = habit.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Medium
                )
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(Spacing.xs)
                ) {
                    // Period badge
                    Surface(
                        shape = RoundedCornerShape(50),
                        color = MaterialTheme.colorScheme.surfaceVariant
                    ) {
                        Text(
                            text = habit.period.title,
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(horizontal = Spacing.xs, vertical = 2.dp)
                        )
                    }
                    // Streak badge
                    if (habit.streak > 0) {
                        Surface(
                            shape = RoundedCornerShape(50),
                            color = Color(0xFFFF9800).copy(alpha = 0.15f)
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = Spacing.xs, vertical = 2.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(2.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.LocalFireDepartment,
                                    contentDescription = null,
                                    modifier = Modifier.size(12.dp),
                                    tint = Color(0xFFFF9800)
                                )
                                Text(
                                    text = "${habit.streak}",
                                    style = MaterialTheme.typography.labelSmall,
                                    fontWeight = FontWeight.SemiBold,
                                    color = Color(0xFFFF9800)
                                )
                            }
                        }
                    }
                }
            }

            // Animated Toggle Button
            AnimatedCheckmark(
                checked = isCompletedOnDate,
                onToggle = onToggle,
                size = 36.dp,
                checkedColor = habit.habitColor
            )
        }

        DropdownMenu(
            expanded = showMenu,
            onDismissRequest = { showMenu = false }
        ) {
            DropdownMenuItem(
                text = { Text("Archive") },
                onClick = {
                    showMenu = false
                    onArchive()
                },
                leadingIcon = {
                    Icon(Icons.Outlined.Archive, contentDescription = null)
                }
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun ArchivedHabitsSheet(
    habits: List<Habit>,
    onDismiss: () -> Unit,
    onUnarchive: (Habit) -> Unit,
    onDelete: (Habit) -> Unit
) {
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Archived Habits",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(16.dp))

            habits.forEach { habit ->
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 4.dp),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .size(40.dp)
                                .clip(RoundedCornerShape(10.dp))
                                .background(habit.habitColor.copy(alpha = 0.15f)),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = habit.icon.toEmojiOrDefault(),
                                style = MaterialTheme.typography.titleMedium
                            )
                        }

                        Spacer(modifier = Modifier.width(12.dp))

                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = habit.title,
                                style = MaterialTheme.typography.bodyLarge
                            )
                            Text(
                                text = habit.period.title,
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }

                        IconButton(onClick = { onUnarchive(habit) }) {
                            Icon(
                                Icons.Default.Unarchive,
                                contentDescription = "Restore",
                                tint = Primary
                            )
                        }

                        IconButton(onClick = { onDelete(habit) }) {
                            Icon(
                                Icons.Default.Delete,
                                contentDescription = "Delete",
                                tint = MaterialTheme.colorScheme.error
                            )
                        }
                    }
                }
            }

            if (habits.isEmpty()) {
                Text(
                    text = "No archived habits",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(vertical = 24.dp)
                )
            }

            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
fun EmptyHabitsView(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = "🔄",
            style = MaterialTheme.typography.displayLarge
        )
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "No habits yet",
            style = MaterialTheme.typography.titleMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            text = "Tap + to create your first habit",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

val Habit.habitColor: Color
    get() = when (color.lowercase()) {
        "green" -> Color(0xFF22C55E)
        "blue" -> Color(0xFF3B82F6)
        "purple" -> Color(0xFF8B5CF6)
        "red" -> Color(0xFFEF4444)
        "orange" -> Color(0xFFF59E0B)
        "pink" -> Color(0xFFEC4899)
        else -> Color(0xFF22C55E)
    }

fun String.toEmojiOrDefault(): String {
    return if (this.length <= 2) this else "✨"
}

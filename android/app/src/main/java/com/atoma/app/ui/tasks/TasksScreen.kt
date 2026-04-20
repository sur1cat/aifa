package com.atoma.app.ui.tasks

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
import androidx.compose.material.icons.filled.CalendarToday
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.ChevronLeft
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.atoma.app.R
import com.atoma.app.domain.model.DailyTask
import com.atoma.app.domain.model.TaskPriority
import com.atoma.app.ui.components.AnimatedCheckmark
import com.atoma.app.ui.theme.CornerRadius
import com.atoma.app.ui.theme.Primary
import com.atoma.app.ui.theme.Spacing
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.format.TextStyle
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TasksScreen(
    viewModel: TasksViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(
                            text = stringResource(R.string.tasks_title),
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            text = stringResource(R.string.tasks_subtitle),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { viewModel.showAddDialog() },
                containerColor = MaterialTheme.colorScheme.primary
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.add_task))
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Date Selector Strip
            DateSelectorStrip(
                selectedDate = uiState.selectedDate,
                onDateSelected = { viewModel.setDate(it) }
            )

            Box(modifier = Modifier.weight(1f)) {
                when {
                    uiState.isLoading -> {
                        CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    }
                    uiState.tasks.isEmpty() -> {
                        EmptyTasksView(modifier = Modifier.align(Alignment.Center))
                    }
                    else -> {
                        LazyColumn(
                            modifier = Modifier.fillMaxSize(),
                            contentPadding = PaddingValues(Spacing.md),
                            verticalArrangement = Arrangement.spacedBy(Spacing.sm)
                        ) {
                            // Progress Card
                            item {
                                TasksProgressCard(
                                    completedCount = uiState.tasks.count { it.isCompleted },
                                    totalCount = uiState.tasks.size
                                )
                            }

                            // Tasks
                            items(uiState.tasks, key = { it.id }) { task ->
                                TaskCard(
                                    task = task,
                                    onToggle = { viewModel.toggleTask(task) },
                                    onClick = { /* TODO */ }
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    if (uiState.showAddDialog) {
        AddTaskDialog(
            initialDate = uiState.selectedDate,
            onDismiss = { viewModel.hideAddDialog() },
            onConfirm = { title, priority, dueDate ->
                viewModel.createTask(title, priority, dueDate)
            }
        )
    }
}

@Composable
fun TaskCard(
    task: DailyTask,
    onToggle: () -> Unit,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier
            .fillMaxWidth()
            .clickable { onClick() },
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
            // Animated Toggle Button
            AnimatedCheckmark(
                checked = task.isCompleted,
                onToggle = onToggle,
                size = 28.dp
            )

            Spacer(modifier = Modifier.width(Spacing.md))

            // Title
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = task.title,
                    style = MaterialTheme.typography.bodyLarge,
                    textDecoration = if (task.isCompleted) TextDecoration.LineThrough else null,
                    color = if (task.isCompleted)
                        MaterialTheme.colorScheme.onSurfaceVariant
                    else
                        MaterialTheme.colorScheme.onSurface
                )
            }

            Spacer(modifier = Modifier.width(Spacing.xs))

            // Priority Badge
            Surface(
                shape = RoundedCornerShape(50),
                color = task.priorityColor.copy(alpha = 0.15f)
            ) {
                Text(
                    text = task.priority.title,
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Medium,
                    color = task.priorityColor,
                    modifier = Modifier.padding(horizontal = Spacing.xs, vertical = 2.dp)
                )
            }
        }
    }
}

@Composable
fun EmptyTasksView(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(Spacing.xxl),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Icon(
            imageVector = Icons.Default.CheckCircle,
            contentDescription = null,
            modifier = Modifier.size(48.dp),
            tint = Primary.copy(alpha = 0.6f)
        )
        Spacer(modifier = Modifier.height(Spacing.md))
        Text(
            text = "All done!",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = MaterialTheme.colorScheme.onSurface
        )
        Spacer(modifier = Modifier.height(Spacing.xxs))
        Text(
            text = "No tasks for this day",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Spacer(modifier = Modifier.height(Spacing.xxs))
        Text(
            text = "Tap + to add a task",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun DateSelectorStrip(
    selectedDate: LocalDate,
    onDateSelected: (LocalDate) -> Unit
) {
    val today = LocalDate.now()
    val dates = remember(today) {
        (-3..7).map { today.plusDays(it.toLong()) }
    }
    val dateFormatter = remember { DateTimeFormatter.ofPattern("d") }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surface)
    ) {
        // Month/Year Header
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            IconButton(onClick = {
                onDateSelected(selectedDate.minusDays(7))
            }) {
                Icon(
                    Icons.Default.ChevronLeft,
                    contentDescription = "Previous week",
                    tint = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Text(
                text = selectedDate.format(DateTimeFormatter.ofPattern("MMMM yyyy")),
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )

            IconButton(onClick = {
                onDateSelected(selectedDate.plusDays(7))
            }) {
                Icon(
                    Icons.Default.ChevronRight,
                    contentDescription = "Next week",
                    tint = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }

        // Date Chips
        LazyRow(
            contentPadding = PaddingValues(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(dates) { date ->
                val isSelected = date == selectedDate
                val isToday = date == today

                Surface(
                    onClick = { onDateSelected(date) },
                    shape = RoundedCornerShape(12.dp),
                    color = when {
                        isSelected -> Primary
                        else -> MaterialTheme.colorScheme.surfaceVariant
                    }
                ) {
                    Column(
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = date.dayOfWeek.getDisplayName(TextStyle.SHORT, Locale.getDefault()),
                            style = MaterialTheme.typography.labelSmall,
                            color = when {
                                isSelected -> Color.White
                                else -> MaterialTheme.colorScheme.onSurfaceVariant
                            }
                        )
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = date.format(dateFormatter),
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                            color = when {
                                isSelected -> Color.White
                                else -> MaterialTheme.colorScheme.onSurface
                            }
                        )
                        if (isToday && !isSelected) {
                            Spacer(modifier = Modifier.height(2.dp))
                            Box(
                                modifier = Modifier
                                    .size(4.dp)
                                    .clip(CircleShape)
                                    .background(Primary)
                            )
                        }
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(8.dp))
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AddTaskDialog(
    initialDate: LocalDate,
    onDismiss: () -> Unit,
    onConfirm: (title: String, priority: TaskPriority, dueDate: LocalDate) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var selectedPriority by remember { mutableStateOf(TaskPriority.MEDIUM) }
    var selectedDate by remember { mutableStateOf(initialDate) }
    var showDatePicker by remember { mutableStateOf(false) }

    val datePickerState = rememberDatePickerState(
        initialSelectedDateMillis = selectedDate.toEpochDay() * 24 * 60 * 60 * 1000
    )
    val dateFormatter = remember { DateTimeFormatter.ofPattern("MMM d, yyyy") }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.add_task)) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text("Task") },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true
                )

                // Due Date
                Text(
                    text = "Due Date",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Surface(
                    onClick = { showDatePicker = true },
                    shape = RoundedCornerShape(12.dp),
                    color = MaterialTheme.colorScheme.surfaceVariant
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp),
                        horizontalArrangement = Arrangement.spacedBy(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            Icons.Default.CalendarToday,
                            contentDescription = null,
                            modifier = Modifier.size(20.dp),
                            tint = Primary
                        )
                        Text(
                            text = selectedDate.format(dateFormatter),
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                }

                // Priority
                Text(
                    text = "Priority",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    TaskPriority.entries.forEach { priority ->
                        FilterChip(
                            selected = priority == selectedPriority,
                            onClick = { selectedPriority = priority },
                            label = { Text(priority.title) },
                            colors = FilterChipDefaults.filterChipColors(
                                selectedContainerColor = when (priority) {
                                    TaskPriority.URGENT -> Color(0xFFDC2626)
                                    TaskPriority.HIGH -> Color(0xFFEF4444)
                                    TaskPriority.MEDIUM -> Color(0xFFF59E0B)
                                    TaskPriority.LOW -> Color(0xFF22C55E)
                                }.copy(alpha = 0.2f),
                                selectedLabelColor = when (priority) {
                                    TaskPriority.URGENT -> Color(0xFFDC2626)
                                    TaskPriority.HIGH -> Color(0xFFEF4444)
                                    TaskPriority.MEDIUM -> Color(0xFFF59E0B)
                                    TaskPriority.LOW -> Color(0xFF22C55E)
                                }
                            )
                        )
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = { if (title.isNotBlank()) onConfirm(title, selectedPriority, selectedDate) },
                enabled = title.isNotBlank()
            ) {
                Text(stringResource(R.string.save))
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(stringResource(R.string.cancel))
            }
        }
    )

    if (showDatePicker) {
        DatePickerDialog(
            onDismissRequest = { showDatePicker = false },
            confirmButton = {
                TextButton(
                    onClick = {
                        datePickerState.selectedDateMillis?.let { millis ->
                            selectedDate = LocalDate.ofEpochDay(millis / (24 * 60 * 60 * 1000))
                        }
                        showDatePicker = false
                    }
                ) {
                    Text("OK")
                }
            },
            dismissButton = {
                TextButton(onClick = { showDatePicker = false }) {
                    Text("Cancel")
                }
            }
        ) {
            DatePicker(state = datePickerState)
        }
    }
}

val DailyTask.priorityColor: Color
    get() = when (priority) {
        TaskPriority.URGENT -> Color(0xFFDC2626)
        TaskPriority.HIGH -> Color(0xFFEF4444)
        TaskPriority.MEDIUM -> Color(0xFFF59E0B)
        TaskPriority.LOW -> Color(0xFF22C55E)
    }

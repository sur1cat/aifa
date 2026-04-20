package com.atoma.app.ui.budget

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.atoma.app.R
import com.atoma.app.domain.model.RecurrenceFrequency
import com.atoma.app.domain.model.RecurringCategory
import com.atoma.app.domain.model.RecurringTransaction
import com.atoma.app.domain.model.Transaction
import com.atoma.app.domain.model.TransactionCategory
import com.atoma.app.domain.model.TransactionType
import com.atoma.app.ui.components.CollapsibleSectionHeader
import com.atoma.app.ui.theme.*
import java.text.NumberFormat
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BudgetScreen(
    budgetViewModel: BudgetViewModel = hiltViewModel(),
    recurringViewModel: RecurringViewModel = hiltViewModel(),
    forecastViewModel: ForecastViewModel = hiltViewModel(),
    savingsGoalViewModel: SavingsGoalViewModel = hiltViewModel()
) {
    val budgetState by budgetViewModel.uiState.collectAsState()
    val recurringState by recurringViewModel.uiState.collectAsState()
    val forecastState by forecastViewModel.uiState.collectAsState()
    val savingsGoalState by savingsGoalViewModel.uiState.collectAsState()
    val currencyFormatter = remember { NumberFormat.getCurrencyInstance(Locale.US) }

    var showRecurringSheet by remember { mutableStateOf(false) }
    var transactionsExpanded by remember { mutableStateOf(true) }
    var recurringExpanded by remember { mutableStateOf(true) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(
                            text = stringResource(R.string.budget_title),
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            text = stringResource(R.string.budget_subtitle),
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
                onClick = { budgetViewModel.showAddDialog() },
                containerColor = MaterialTheme.colorScheme.primary
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.add_transaction))
            }
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(Spacing.md),
            verticalArrangement = Arrangement.spacedBy(Spacing.sm)
        ) {
            // Summary Card
            item {
                SummaryCard(
                    income = budgetState.summary.income,
                    expenses = budgetState.summary.expenses,
                    balance = budgetState.summary.balance,
                    formatter = currencyFormatter
                )
            }

            // Savings Goal Card
            item {
                SavingsGoalCard(
                    savingsGoal = savingsGoalState.savingsGoal,
                    monthlyIncome = savingsGoalState.monthlyIncome,
                    monthlyExpenses = savingsGoalState.monthlyExpenses,
                    onEdit = { savingsGoalViewModel.showEditSheet() },
                    onSetGoal = { savingsGoalViewModel.showEditSheet() }
                )
            }

            // Budget Forecast Card
            forecastState.forecast?.let { forecast ->
                item {
                    BudgetForecastCard(forecast = forecast)
                }
            }

            // Category Breakdown Card
            if (budgetState.transactions.isNotEmpty()) {
                item {
                    CategoryBreakdownCard(transactions = budgetState.transactions)
                }
            }

            // Recurring Section with Collapsible Header
            item {
                CollapsibleSectionHeader(
                    title = "Recurring",
                    expanded = recurringExpanded,
                    onToggle = { recurringExpanded = !recurringExpanded },
                    icon = Icons.Default.Sync,
                    count = recurringState.activeTransactions.size
                )
            }

            item {
                AnimatedVisibility(
                    visible = recurringExpanded,
                    enter = expandVertically() + fadeIn(),
                    exit = shrinkVertically() + fadeOut()
                ) {
                    RecurringSectionContent(
                        transactions = recurringState.activeTransactions,
                        monthlyExpenses = recurringState.monthlyExpenses,
                        formatter = currencyFormatter,
                        onViewAll = { showRecurringSheet = true },
                        onAdd = { recurringViewModel.showAddSheet() }
                    )
                }
            }

            // Transactions Section with Collapsible Header
            item {
                CollapsibleSectionHeader(
                    title = "Transactions",
                    expanded = transactionsExpanded,
                    onToggle = { transactionsExpanded = !transactionsExpanded },
                    icon = Icons.Default.Receipt,
                    count = budgetState.transactions.size
                )
            }

            // Transactions Content
            if (budgetState.transactions.isEmpty() && !budgetState.isLoading) {
                item {
                    AnimatedVisibility(
                        visible = transactionsExpanded,
                        enter = expandVertically() + fadeIn(),
                        exit = shrinkVertically() + fadeOut()
                    ) {
                        EmptyBudgetView()
                    }
                }
            } else {
                item {
                    AnimatedVisibility(
                        visible = transactionsExpanded,
                        enter = expandVertically() + fadeIn(),
                        exit = shrinkVertically() + fadeOut()
                    ) {
                        Column(verticalArrangement = Arrangement.spacedBy(Spacing.xs)) {
                            budgetState.transactions.forEach { transaction ->
                                TransactionCard(
                                    transaction = transaction,
                                    formatter = currencyFormatter,
                                    onClick = { /* TODO */ }
                                )
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

    if (budgetState.showAddDialog) {
        AddTransactionDialog(
            onDismiss = { budgetViewModel.hideAddDialog() },
            onConfirm = { title, amount, type, category ->
                budgetViewModel.createTransaction(title, amount, type, category)
            }
        )
    }

    if (recurringState.showAddSheet) {
        AddRecurringSheet(
            onDismiss = { recurringViewModel.hideAddSheet() },
            onConfirm = { title, amount, type, category, frequency ->
                recurringViewModel.createRecurring(title, amount, type, category, frequency)
            }
        )
    }

    if (showRecurringSheet) {
        RecurringListSheet(
            transactions = recurringState.transactions,
            formatter = currencyFormatter,
            onDismiss = { showRecurringSheet = false },
            onToggle = { recurringViewModel.toggleActive(it) },
            onDelete = { recurringViewModel.deleteRecurring(it) }
        )
    }

    if (savingsGoalState.showEditSheet) {
        EditSavingsGoalSheet(
            currentGoal = savingsGoalState.savingsGoal,
            onDismiss = { savingsGoalViewModel.hideEditSheet() },
            onSave = { target -> savingsGoalViewModel.saveGoal(target) },
            onDelete = { savingsGoalViewModel.deleteGoal() }
        )
    }
}

@Composable
private fun RecurringSectionContent(
    transactions: List<RecurringTransaction>,
    monthlyExpenses: Double,
    formatter: NumberFormat,
    onViewAll: () -> Unit,
    onAdd: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(CornerRadius.large),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
        )
    ) {
        Column(modifier = Modifier.padding(Spacing.md)) {
            if (transactions.isEmpty()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { onAdd() }
                        .padding(vertical = Spacing.sm),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        Icons.Default.Add,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Spacer(modifier = Modifier.width(Spacing.xs))
                    Text(
                        text = "Add recurring payment",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            } else {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = "${transactions.size} active",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = "${formatter.format(monthlyExpenses)}/mo",
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Expense
                    )
                }

                Spacer(modifier = Modifier.height(Spacing.xs))

                // Show first 3 recurring
                transactions.take(3).forEach { recurring ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = Spacing.xxs),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = recurring.title,
                            style = MaterialTheme.typography.bodyMedium,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                            modifier = Modifier.weight(1f)
                        )
                        Text(
                            text = formatter.format(recurring.amount),
                            style = MaterialTheme.typography.bodyMedium,
                            color = if (recurring.type == TransactionType.INCOME) Income else Expense
                        )
                    }
                }

                if (transactions.size > 3) {
                    TextButton(
                        onClick = onViewAll,
                        modifier = Modifier.padding(top = Spacing.xxs)
                    ) {
                        Text("+${transactions.size - 3} more - View all")
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun AddRecurringSheet(
    onDismiss: () -> Unit,
    onConfirm: (String, Double, TransactionType, RecurringCategory, RecurrenceFrequency) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var amount by remember { mutableStateOf("") }
    var selectedType by remember { mutableStateOf(TransactionType.EXPENSE) }
    var selectedCategory by remember { mutableStateOf(RecurringCategory.SUBSCRIPTIONS) }
    var selectedFrequency by remember { mutableStateOf(RecurrenceFrequency.MONTHLY) }

    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Text(
                text = "Add Recurring",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )

            OutlinedTextField(
                value = title,
                onValueChange = { title = it },
                label = { Text("Description") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )

            OutlinedTextField(
                value = amount,
                onValueChange = { amount = it.filter { c -> c.isDigit() || c == '.' } },
                label = { Text("Amount") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )

            // Type
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                FilterChip(
                    selected = selectedType == TransactionType.EXPENSE,
                    onClick = { selectedType = TransactionType.EXPENSE },
                    label = { Text("Expense") },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = Expense.copy(alpha = 0.15f),
                        selectedLabelColor = Expense
                    )
                )
                FilterChip(
                    selected = selectedType == TransactionType.INCOME,
                    onClick = { selectedType = TransactionType.INCOME },
                    label = { Text("Income") },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = Income.copy(alpha = 0.15f),
                        selectedLabelColor = Income
                    )
                )
            }

            // Frequency
            Text(
                text = "Frequency",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(RecurrenceFrequency.entries) { freq ->
                    FilterChip(
                        selected = selectedFrequency == freq,
                        onClick = { selectedFrequency = freq },
                        label = { Text(freq.title) }
                    )
                }
            }

            // Category
            Text(
                text = "Category",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(RecurringCategory.entries) { cat ->
                    FilterChip(
                        selected = selectedCategory == cat,
                        onClick = { selectedCategory = cat },
                        label = { Text(cat.name.lowercase().replaceFirstChar { it.uppercase() }) }
                    )
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            Button(
                onClick = {
                    val amountValue = amount.toDoubleOrNull() ?: 0.0
                    if (title.isNotBlank() && amountValue > 0) {
                        onConfirm(title, amountValue, selectedType, selectedCategory, selectedFrequency)
                    }
                },
                modifier = Modifier.fillMaxWidth(),
                enabled = title.isNotBlank() && amount.isNotBlank()
            ) {
                Text("Add")
            }

            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RecurringListSheet(
    transactions: List<RecurringTransaction>,
    formatter: NumberFormat,
    onDismiss: () -> Unit,
    onToggle: (RecurringTransaction) -> Unit,
    onDelete: (RecurringTransaction) -> Unit
) {
    ModalBottomSheet(onDismissRequest = onDismiss) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Recurring Transactions",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(16.dp))

            transactions.forEach { recurring ->
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
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = recurring.title,
                                style = MaterialTheme.typography.bodyLarge,
                                fontWeight = FontWeight.Medium,
                                color = if (recurring.isActive) {
                                    MaterialTheme.colorScheme.onSurface
                                } else {
                                    MaterialTheme.colorScheme.onSurfaceVariant
                                }
                            )
                            Text(
                                text = "${formatter.format(recurring.amount)} / ${recurring.frequency.title.lowercase()}",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }

                        Switch(
                            checked = recurring.isActive,
                            onCheckedChange = { onToggle(recurring) }
                        )

                        IconButton(onClick = { onDelete(recurring) }) {
                            Icon(
                                Icons.Default.Delete,
                                contentDescription = "Delete",
                                tint = MaterialTheme.colorScheme.error
                            )
                        }
                    }
                }
            }

            if (transactions.isEmpty()) {
                Text(
                    text = "No recurring transactions",
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
fun SummaryCard(
    income: Double,
    expenses: Double,
    balance: Double,
    formatter: NumberFormat
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Column(
            modifier = Modifier.padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Balance
            Column {
                Text(
                    text = "Balance",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = formatter.format(balance),
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold,
                    color = if (balance >= 0) Income else Expense
                )
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                // Income
                Column {
                    Text(
                        text = "Income",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = formatter.format(income),
                        style = MaterialTheme.typography.titleMedium,
                        color = Income,
                        fontWeight = FontWeight.SemiBold
                    )
                }

                // Expenses
                Column(horizontalAlignment = Alignment.End) {
                    Text(
                        text = "Expenses",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = formatter.format(expenses),
                        style = MaterialTheme.typography.titleMedium,
                        color = Expense,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

@Composable
fun TransactionCard(
    transaction: Transaction,
    formatter: NumberFormat,
    onClick: () -> Unit
) {
    val category = transaction.categoryEnum
    val categoryColor = category.color

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Category Icon
            Box(
                modifier = Modifier
                    .size(44.dp)
                    .clip(CircleShape)
                    .background(categoryColor.copy(alpha = 0.15f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = category.iconVector,
                    contentDescription = null,
                    tint = categoryColor,
                    modifier = Modifier.size(22.dp)
                )
            }

            Spacer(modifier = Modifier.width(16.dp))

            // Title & Category
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = transaction.title,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.Medium
                )
                Text(
                    text = stringResource(category.titleRes),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            // Amount
            Text(
                text = "${if (transaction.type == TransactionType.INCOME) "+" else "-"}${formatter.format(transaction.amount)}",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                color = if (transaction.type == TransactionType.INCOME) Income else Expense
            )
        }
    }
}

@Composable
fun EmptyBudgetView() {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(text = "💰", style = MaterialTheme.typography.displayLarge)
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "No transactions yet",
            style = MaterialTheme.typography.titleMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AddTransactionDialog(
    onDismiss: () -> Unit,
    onConfirm: (title: String, amount: Double, type: TransactionType, category: String) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var amount by remember { mutableStateOf("") }
    var selectedType by remember { mutableStateOf(TransactionType.EXPENSE) }
    var selectedCategory by remember { mutableStateOf(TransactionCategory.FOOD) }

    val availableCategories = if (selectedType == TransactionType.INCOME) {
        TransactionCategory.incomeCategories
    } else {
        TransactionCategory.expenseCategories
    }

    // Reset category when type changes
    LaunchedEffect(selectedType) {
        selectedCategory = if (selectedType == TransactionType.INCOME) {
            TransactionCategory.SALARY
        } else {
            TransactionCategory.FOOD
        }
    }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.add_transaction)) },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text("Description") },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true
                )

                OutlinedTextField(
                    value = amount,
                    onValueChange = { amount = it.filter { c -> c.isDigit() || c == '.' } },
                    label = { Text("Amount") },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    singleLine = true
                )

                // Type selector
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    FilterChip(
                        selected = selectedType == TransactionType.EXPENSE,
                        onClick = { selectedType = TransactionType.EXPENSE },
                        label = { Text(stringResource(R.string.expense)) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = Expense.copy(alpha = 0.15f),
                            selectedLabelColor = Expense
                        )
                    )
                    FilterChip(
                        selected = selectedType == TransactionType.INCOME,
                        onClick = { selectedType = TransactionType.INCOME },
                        label = { Text(stringResource(R.string.income)) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = Income.copy(alpha = 0.15f),
                            selectedLabelColor = Income
                        )
                    )
                }

                // Category label
                Text(
                    text = stringResource(R.string.category),
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                // Category grid
                LazyRow(
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    items(availableCategories) { category ->
                        CategoryChip(
                            category = category,
                            isSelected = category == selectedCategory,
                            onClick = { selectedCategory = category }
                        )
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    val amountValue = amount.toDoubleOrNull() ?: 0.0
                    if (title.isNotBlank() && amountValue > 0) {
                        onConfirm(title, amountValue, selectedType, selectedCategory.key)
                    }
                },
                enabled = title.isNotBlank() && amount.isNotBlank()
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
}

@Composable
fun CategoryChip(
    category: TransactionCategory,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    val categoryColor = category.color

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clickable { onClick() }
            .padding(4.dp)
    ) {
        Box(
            modifier = Modifier
                .size(48.dp)
                .clip(CircleShape)
                .background(
                    if (isSelected) categoryColor else categoryColor.copy(alpha = 0.15f)
                )
                .then(
                    if (isSelected) Modifier.border(2.dp, categoryColor, CircleShape)
                    else Modifier
                ),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = category.iconVector,
                contentDescription = null,
                tint = if (isSelected) Color.White else categoryColor,
                modifier = Modifier.size(22.dp)
            )
        }

        Spacer(modifier = Modifier.height(4.dp))

        Text(
            text = stringResource(category.titleRes),
            style = MaterialTheme.typography.labelSmall,
            color = if (isSelected) {
                MaterialTheme.colorScheme.onSurface
            } else {
                MaterialTheme.colorScheme.onSurfaceVariant
            },
            maxLines = 1
        )
    }
}

// Extension properties for TransactionCategory
val TransactionCategory.color: Color
    get() = when (this) {
        TransactionCategory.FOOD -> CategoryFood
        TransactionCategory.TRANSPORT -> CategoryTransport
        TransactionCategory.SHOPPING -> CategoryShopping
        TransactionCategory.ENTERTAINMENT -> CategoryEntertainment
        TransactionCategory.HEALTH -> CategoryHealth
        TransactionCategory.EDUCATION -> CategoryEducation
        TransactionCategory.BILLS -> CategoryBills
        TransactionCategory.SALARY -> CategorySalary
        TransactionCategory.FREELANCE -> CategoryFreelance
        TransactionCategory.INVESTMENT -> CategoryInvestment
        TransactionCategory.GIFT -> CategoryGift
        TransactionCategory.OTHER -> CategoryOther
    }

val TransactionCategory.iconVector: ImageVector
    get() = when (this) {
        TransactionCategory.FOOD -> Icons.Outlined.Restaurant
        TransactionCategory.TRANSPORT -> Icons.Outlined.DirectionsCar
        TransactionCategory.SHOPPING -> Icons.Outlined.ShoppingBag
        TransactionCategory.ENTERTAINMENT -> Icons.Outlined.Movie
        TransactionCategory.HEALTH -> Icons.Outlined.Favorite
        TransactionCategory.EDUCATION -> Icons.Outlined.School
        TransactionCategory.BILLS -> Icons.Outlined.Receipt
        TransactionCategory.SALARY -> Icons.Outlined.Payments
        TransactionCategory.FREELANCE -> Icons.Outlined.Laptop
        TransactionCategory.INVESTMENT -> Icons.Outlined.TrendingUp
        TransactionCategory.GIFT -> Icons.Outlined.Redeem
        TransactionCategory.OTHER -> Icons.Outlined.MoreHoriz
    }

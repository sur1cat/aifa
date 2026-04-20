package com.atoma.app.ui.budget

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PieChart
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.atoma.app.domain.model.Transaction
import com.atoma.app.domain.model.TransactionCategory
import com.atoma.app.domain.model.TransactionType
import com.atoma.app.ui.theme.Primary
import java.text.NumberFormat
import java.util.Locale

data class CategorySpending(
    val category: TransactionCategory,
    val amount: Double,
    val percentage: Float
)

@Composable
fun CategoryBreakdownCard(
    transactions: List<Transaction>,
    modifier: Modifier = Modifier
) {
    val currencyFormatter = remember { NumberFormat.getCurrencyInstance(Locale.US) }

    // Calculate spending by category
    val categorySpending = remember(transactions) {
        val expenses = transactions.filter { it.type == TransactionType.EXPENSE }
        val totalExpenses = expenses.sumOf { it.amount }

        if (totalExpenses == 0.0) return@remember emptyList()

        expenses
            .groupBy { it.categoryEnum }
            .map { (category, txs) ->
                val amount = txs.sumOf { it.amount }
                CategorySpending(
                    category = category,
                    amount = amount,
                    percentage = (amount / totalExpenses).toFloat()
                )
            }
            .sortedByDescending { it.amount }
            .take(5)
    }

    Card(
        modifier = modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            // Header
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.PieChart,
                    contentDescription = null,
                    modifier = Modifier.size(18.dp),
                    tint = Primary
                )
                Text(
                    text = "Spending by Category",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            if (categorySpending.isEmpty()) {
                Text(
                    text = "No expenses this month",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(vertical = 16.dp)
                )
            } else {
                categorySpending.forEach { spending ->
                    CategoryProgressRow(
                        spending = spending,
                        formatter = currencyFormatter
                    )
                    Spacer(modifier = Modifier.height(12.dp))
                }
            }
        }
    }
}

@Composable
private fun CategoryProgressRow(
    spending: CategorySpending,
    formatter: NumberFormat
) {
    val animatedProgress by animateFloatAsState(
        targetValue = spending.percentage,
        animationSpec = tween(durationMillis = 800),
        label = "progress"
    )

    val categoryColor = spending.category.color

    Column {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Box(
                    modifier = Modifier
                        .size(32.dp)
                        .clip(CircleShape)
                        .background(categoryColor.copy(alpha = 0.15f)),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = spending.category.iconVector,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp),
                        tint = categoryColor
                    )
                }

                Text(
                    text = spending.category.name.lowercase().replaceFirstChar { it.uppercase() },
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium
                )
            }

            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = formatter.format(spending.amount),
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    text = "${(spending.percentage * 100).toInt()}%",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }

        Spacer(modifier = Modifier.height(6.dp))

        // Progress bar
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(6.dp)
                .clip(RoundedCornerShape(3.dp))
                .background(categoryColor.copy(alpha = 0.15f))
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(animatedProgress)
                    .fillMaxHeight()
                    .clip(RoundedCornerShape(3.dp))
                    .background(categoryColor)
            )
        }
    }
}

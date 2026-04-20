package com.atoma.app.ui.budget

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.TrendingDown
import androidx.compose.material.icons.automirrored.filled.TrendingFlat
import androidx.compose.material.icons.automirrored.filled.TrendingUp
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.atoma.app.domain.model.*
import com.atoma.app.ui.theme.Primary
import java.time.format.DateTimeFormatter

@Composable
fun BudgetForecastCard(
    forecast: BudgetForecast,
    modifier: Modifier = Modifier
) {
    var isExpanded by remember { mutableStateOf(false) }

    val monthFormatter = remember { DateTimeFormatter.ofPattern("MMMM yyyy") }

    Card(
        modifier = modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        ),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column {
            // Header
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { isExpanded = !isExpanded }
                    .padding(16.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(6.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.ShowChart,
                            contentDescription = null,
                            modifier = Modifier.size(16.dp),
                            tint = Primary
                        )
                        Text(
                            text = "Next Month Forecast",
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                    Spacer(modifier = Modifier.height(2.dp))
                    Text(
                        text = forecast.forecastMonth.format(monthFormatter),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    Text(
                        text = "${(forecast.confidenceScore * 100).toInt()}%",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Icon(
                        imageVector = if (isExpanded) Icons.Default.ExpandLess else Icons.Default.ExpandMore,
                        contentDescription = if (isExpanded) "Collapse" else "Expand",
                        modifier = Modifier.size(20.dp),
                        tint = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))

            // Summary Stats
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(vertical = 12.dp),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                ForecastStatItem(
                    title = "Expenses",
                    amount = forecast.projectedExpenses,
                    trend = forecast.expenseTrend,
                    color = MaterialTheme.colorScheme.error
                )

                Box(
                    modifier = Modifier
                        .width(1.dp)
                        .height(40.dp)
                        .background(MaterialTheme.colorScheme.outlineVariant)
                )

                ForecastStatItem(
                    title = "Income",
                    amount = forecast.projectedIncome,
                    trend = TrendDirection.STABLE,
                    color = Color(0xFF4CAF50)
                )

                Box(
                    modifier = Modifier
                        .width(1.dp)
                        .height(40.dp)
                        .background(MaterialTheme.colorScheme.outlineVariant)
                )

                ForecastStatItem(
                    title = "Savings",
                    amount = forecast.projectedSavings,
                    trend = if (forecast.projectedSavings >= 0) TrendDirection.UP else TrendDirection.DOWN,
                    color = if (forecast.projectedSavings >= 0) Color(0xFF4CAF50) else MaterialTheme.colorScheme.error
                )
            }

            // Seasonal Warnings
            forecast.seasonalFactors?.let { factors ->
                if (factors.isNotEmpty()) {
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))

                    Column(
                        modifier = Modifier.padding(12.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        factors.forEach { factor ->
                            SeasonalWarningRow(factor = factor)
                        }
                    }
                }
            }

            // Expandable Category Breakdown
            AnimatedVisibility(
                visible = isExpanded && forecast.categoryForecasts.isNotEmpty(),
                enter = expandVertically(),
                exit = shrinkVertically()
            ) {
                Column {
                    HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp))

                    forecast.categoryForecasts.forEachIndexed { index, category ->
                        CategoryForecastRow(forecast = category)
                        if (index < forecast.categoryForecasts.size - 1) {
                            HorizontalDivider(
                                modifier = Modifier.padding(start = 60.dp, end = 16.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(8.dp))
                }
            }
        }
    }
}

@Composable
private fun ForecastStatItem(
    title: String,
    amount: Double,
    trend: TrendDirection,
    color: Color
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Spacer(modifier = Modifier.height(4.dp))
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(2.dp)
        ) {
            Text(
                text = formatCurrency(kotlin.math.abs(amount)),
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
                color = color
            )
            if (trend != TrendDirection.STABLE) {
                Icon(
                    imageVector = when (trend) {
                        TrendDirection.UP -> Icons.AutoMirrored.Filled.TrendingUp
                        TrendDirection.DOWN -> Icons.AutoMirrored.Filled.TrendingDown
                        TrendDirection.STABLE -> Icons.AutoMirrored.Filled.TrendingFlat
                    },
                    contentDescription = null,
                    modifier = Modifier.size(14.dp),
                    tint = when (trend) {
                        TrendDirection.UP -> MaterialTheme.colorScheme.error
                        TrendDirection.DOWN -> Color(0xFF4CAF50)
                        TrendDirection.STABLE -> MaterialTheme.colorScheme.onSurfaceVariant
                    }
                )
            }
        }
    }
}

@Composable
private fun SeasonalWarningRow(factor: SeasonalFactor) {
    val increasePercent = ((factor.monthlyMultiplier - 1.0) * 100).toInt()

    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        Icon(
            imageVector = Icons.Default.Warning,
            contentDescription = null,
            modifier = Modifier.size(14.dp),
            tint = Color(0xFFFFA000)
        )

        Text(
            text = factor.category,
            style = MaterialTheme.typography.bodySmall,
            fontWeight = FontWeight.Medium
        )

        Text(
            text = factor.reason,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.weight(1f)
        )

        if (increasePercent > 0) {
            Text(
                text = "+$increasePercent%",
                style = MaterialTheme.typography.labelMedium,
                fontWeight = FontWeight.Medium,
                color = Color(0xFFFFA000)
            )
        }
    }
}

@Composable
private fun CategoryForecastRow(forecast: CategoryForecast) {
    val categoryColor = getCategoryColor(forecast.category)
    val changePercent = if (forecast.historicalAverage > 0) {
        ((forecast.projectedAmount - forecast.historicalAverage) / forecast.historicalAverage * 100).toInt()
    } else 0

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 10.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(32.dp)
                .clip(CircleShape)
                .background(categoryColor.copy(alpha = 0.15f)),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = getCategoryIcon(forecast.category),
                contentDescription = null,
                modifier = Modifier.size(16.dp),
                tint = categoryColor
            )
        }

        Spacer(modifier = Modifier.width(12.dp))

        Column(
            modifier = Modifier.weight(1f)
        ) {
            Text(
                text = forecast.category.replaceFirstChar { it.uppercase() },
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium
            )

            if (forecast.recurringAmount > 0) {
                Text(
                    text = "incl. ${formatCurrency(forecast.recurringAmount)} recurring",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }

        Column(
            horizontalAlignment = Alignment.End
        ) {
            Text(
                text = formatCurrency(forecast.projectedAmount),
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.SemiBold
            )

            if (kotlin.math.abs(changePercent) > 5) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(2.dp)
                ) {
                    Icon(
                        imageVector = if (forecast.trend == TrendDirection.UP)
                            Icons.AutoMirrored.Filled.TrendingUp
                        else
                            Icons.AutoMirrored.Filled.TrendingDown,
                        contentDescription = null,
                        modifier = Modifier.size(12.dp),
                        tint = if (forecast.trend == TrendDirection.UP)
                            MaterialTheme.colorScheme.error
                        else
                            Color(0xFF4CAF50)
                    )
                    Text(
                        text = "${kotlin.math.abs(changePercent)}%",
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.Medium,
                        color = if (forecast.trend == TrendDirection.UP)
                            MaterialTheme.colorScheme.error
                        else
                            Color(0xFF4CAF50)
                    )
                }
            }
        }
    }
}

private fun formatCurrency(amount: Double): String {
    return when {
        amount >= 1_000_000 -> "$${String.format("%.1fM", amount / 1_000_000)}"
        amount >= 1000 -> "$${String.format("%.1fK", amount / 1000)}"
        else -> "$${amount.toInt()}"
    }
}

private fun getCategoryColor(category: String): Color {
    return when (category.lowercase()) {
        "food" -> Color(0xFFFF7043)
        "transport" -> Color(0xFF42A5F5)
        "shopping" -> Color(0xFFAB47BC)
        "entertainment" -> Color(0xFFFFCA28)
        "health" -> Color(0xFFEF5350)
        "education" -> Color(0xFF5C6BC0)
        "bills" -> Color(0xFF78909C)
        "salary" -> Color(0xFF66BB6A)
        "freelance" -> Color(0xFF26A69A)
        "investment" -> Color(0xFF29B6F6)
        "gift" -> Color(0xFFEC407A)
        else -> Color(0xFF9E9E9E)
    }
}

private fun getCategoryIcon(category: String): androidx.compose.ui.graphics.vector.ImageVector {
    return when (category.lowercase()) {
        "food" -> Icons.Default.Restaurant
        "transport" -> Icons.Default.DirectionsCar
        "shopping" -> Icons.Default.ShoppingBag
        "entertainment" -> Icons.Default.Movie
        "health" -> Icons.Default.Favorite
        "education" -> Icons.Default.School
        "bills" -> Icons.Default.Receipt
        "salary" -> Icons.Default.Payments
        "freelance" -> Icons.Default.Laptop
        "investment" -> Icons.AutoMirrored.Filled.TrendingUp
        "gift" -> Icons.Default.CardGiftcard
        else -> Icons.Default.MoreHoriz
    }
}

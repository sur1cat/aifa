package com.atoma.app.ui.habits

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Today
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import com.atoma.app.ui.components.CompactDateSelectorCell
import com.atoma.app.ui.theme.Spacing
import java.time.LocalDate
import java.time.format.DateTimeFormatter

@Composable
fun HabitsDateSelector(
    selectedDate: LocalDate,
    onDateSelected: (LocalDate) -> Unit,
    modifier: Modifier = Modifier
) {
    val today = LocalDate.now()
    val dates = remember(today) {
        (-3..3).map { today.plusDays(it.toLong()) }
    }
    val monthYearFormatter = remember { DateTimeFormatter.ofPattern("MMMM yyyy") }
    val listState = rememberLazyListState()

    // Scroll to center (today) when first displayed
    LaunchedEffect(Unit) {
        listState.scrollToItem(3) // Index 3 is today in the list
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surface)
    ) {
        // Month/Year Header with Today button
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = Spacing.md, vertical = Spacing.xs),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = selectedDate.format(monthYearFormatter),
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onSurface
            )

            if (selectedDate != today) {
                TextButton(onClick = { onDateSelected(today) }) {
                    Icon(
                        imageVector = Icons.Default.Today,
                        contentDescription = "Today",
                        modifier = Modifier.padding(end = Spacing.xxs)
                    )
                    Text("Today")
                }
            }
        }

        // Date Cells
        LazyRow(
            state = listState,
            modifier = Modifier.fillMaxWidth(),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = Spacing.md),
            horizontalArrangement = Arrangement.spacedBy(Spacing.xs)
        ) {
            items(dates) { date ->
                CompactDateSelectorCell(
                    date = date,
                    isSelected = date == selectedDate,
                    isToday = date == today,
                    onClick = { onDateSelected(date) }
                )
            }
        }

        Spacer(modifier = Modifier.height(Spacing.xs))
    }
}

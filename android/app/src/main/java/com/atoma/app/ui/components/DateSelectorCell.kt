package com.atoma.app.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.atoma.app.ui.theme.CornerRadius
import com.atoma.app.ui.theme.Primary
import com.atoma.app.ui.theme.Spacing
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.format.TextStyle
import java.util.Locale

@Composable
fun DateSelectorCell(
    date: LocalDate,
    isSelected: Boolean,
    isToday: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val dayFormatter = remember { DateTimeFormatter.ofPattern("d") }

    val backgroundColor = when {
        isSelected -> Primary
        else -> MaterialTheme.colorScheme.surfaceVariant
    }

    val textColor = when {
        isSelected -> Color.White
        else -> MaterialTheme.colorScheme.onSurface
    }

    val secondaryTextColor = when {
        isSelected -> Color.White.copy(alpha = 0.8f)
        else -> MaterialTheme.colorScheme.onSurfaceVariant
    }

    Surface(
        onClick = onClick,
        shape = RoundedCornerShape(CornerRadius.medium),
        color = backgroundColor,
        modifier = modifier
            .then(
                if (isToday && !isSelected) {
                    Modifier.border(
                        width = 1.5.dp,
                        color = Primary,
                        shape = RoundedCornerShape(CornerRadius.medium)
                    )
                } else Modifier
            )
    ) {
        Column(
            modifier = Modifier.padding(horizontal = Spacing.sm, vertical = Spacing.xs),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = date.dayOfWeek.getDisplayName(TextStyle.SHORT, Locale.getDefault()),
                style = MaterialTheme.typography.labelSmall,
                fontSize = 11.sp,
                color = secondaryTextColor
            )
            Spacer(modifier = Modifier.height(2.dp))
            Text(
                text = date.format(dayFormatter),
                style = MaterialTheme.typography.titleMedium,
                fontSize = 18.sp,
                fontWeight = FontWeight.SemiBold,
                color = textColor
            )
        }
    }
}

@Composable
fun CompactDateSelectorCell(
    date: LocalDate,
    isSelected: Boolean,
    isToday: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val dayFormatter = remember { DateTimeFormatter.ofPattern("d") }

    val backgroundColor = when {
        isSelected -> Primary
        else -> Color.Transparent
    }

    val textColor = when {
        isSelected -> Color.White
        else -> MaterialTheme.colorScheme.onSurface
    }

    val secondaryTextColor = when {
        isSelected -> Color.White.copy(alpha = 0.8f)
        else -> MaterialTheme.colorScheme.onSurfaceVariant
    }

    Surface(
        onClick = onClick,
        shape = RoundedCornerShape(CornerRadius.small),
        color = backgroundColor,
        modifier = modifier.width(44.dp)
    ) {
        Column(
            modifier = Modifier.padding(vertical = Spacing.xs),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = date.dayOfWeek.getDisplayName(TextStyle.NARROW, Locale.getDefault()),
                style = MaterialTheme.typography.labelSmall,
                fontSize = 10.sp,
                color = secondaryTextColor
            )
            Spacer(modifier = Modifier.height(4.dp))
            Box(
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = date.format(dayFormatter),
                    style = MaterialTheme.typography.bodyLarge,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = textColor
                )
            }
            if (isToday && !isSelected) {
                Spacer(modifier = Modifier.height(4.dp))
                Box(
                    modifier = Modifier
                        .size(4.dp)
                        .clip(CircleShape)
                        .background(Primary)
                )
            } else {
                Spacer(modifier = Modifier.height(8.dp))
            }
        }
    }
}

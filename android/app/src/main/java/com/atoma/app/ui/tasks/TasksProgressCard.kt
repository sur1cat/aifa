package com.atoma.app.ui.tasks

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.atoma.app.ui.theme.CornerRadius
import com.atoma.app.ui.theme.Primary
import com.atoma.app.ui.theme.ProgressSize
import com.atoma.app.ui.theme.Spacing

@Composable
fun TasksProgressCard(
    completedCount: Int,
    totalCount: Int,
    modifier: Modifier = Modifier
) {
    val progress = if (totalCount > 0) completedCount.toFloat() / totalCount else 0f
    val animatedProgress by animateFloatAsState(
        targetValue = progress,
        animationSpec = tween(durationMillis = 500),
        label = "progress"
    )

    val percentage = (animatedProgress * 100).toInt()

    Card(
        modifier = modifier.fillMaxWidth(),
        shape = RoundedCornerShape(CornerRadius.large),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(Spacing.md),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Left side: Ring progress
            Box(
                modifier = Modifier.size(ProgressSize.medium),
                contentAlignment = Alignment.Center
            ) {
                // Background ring
                CircularProgressIndicator(
                    progress = { 1f },
                    modifier = Modifier.size(ProgressSize.medium),
                    color = Primary.copy(alpha = 0.15f),
                    strokeWidth = 6.dp,
                    strokeCap = StrokeCap.Round
                )
                // Progress ring
                CircularProgressIndicator(
                    progress = { animatedProgress },
                    modifier = Modifier.size(ProgressSize.medium),
                    color = Primary,
                    strokeWidth = 6.dp,
                    strokeCap = StrokeCap.Round
                )
                // Center text
                Text(
                    text = "$completedCount",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Primary
                )
            }

            // Right side: Text info
            Column(
                horizontalAlignment = Alignment.End
            ) {
                Text(
                    text = "$percentage%",
                    style = MaterialTheme.typography.headlineSmall,
                    fontWeight = FontWeight.Bold,
                    color = Primary
                )
                Text(
                    text = "$completedCount of $totalCount completed",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

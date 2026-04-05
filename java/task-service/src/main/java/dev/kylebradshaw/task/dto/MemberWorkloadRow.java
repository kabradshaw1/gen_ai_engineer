package dev.kylebradshaw.task.dto;

import java.util.UUID;

public record MemberWorkloadRow(UUID userId, String name, int assignedCount, int completedCount) {
}

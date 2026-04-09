package dev.kylebradshaw.gateway.dto;

import java.util.List;
import java.util.Map;

public record ProjectStatsDto(
        TaskStatusCountsDto taskCountByStatus,
        TaskPriorityCountsDto taskCountByPriority,
        int overdueCount,
        Double avgCompletionTimeHours,
        List<MemberWorkloadDto> memberWorkload) {

    public record MemberWorkloadDto(String userId, String name, int assignedCount, int completedCount) {}

    public record TaskStatusCountsDto(int todo, int inProgress, int done) {
        public static TaskStatusCountsDto fromMap(Map<String, Integer> m) {
            if (m == null) {
                return new TaskStatusCountsDto(0, 0, 0);
            }
            return new TaskStatusCountsDto(
                    m.getOrDefault("TODO", 0),
                    m.getOrDefault("IN_PROGRESS", 0),
                    m.getOrDefault("DONE", 0));
        }
    }

    public record TaskPriorityCountsDto(int low, int medium, int high) {
        public static TaskPriorityCountsDto fromMap(Map<String, Integer> m) {
            if (m == null) {
                return new TaskPriorityCountsDto(0, 0, 0);
            }
            return new TaskPriorityCountsDto(
                    m.getOrDefault("LOW", 0),
                    m.getOrDefault("MEDIUM", 0),
                    m.getOrDefault("HIGH", 0));
        }
    }

    /** Raw shape as returned by task-service (maps keyed by DB enum names). */
    public record Raw(
            Map<String, Integer> taskCountByStatus,
            Map<String, Integer> taskCountByPriority,
            int overdueCount,
            Double avgCompletionTimeHours,
            List<MemberWorkloadDto> memberWorkload) {

        public ProjectStatsDto toTyped() {
            return new ProjectStatsDto(
                    TaskStatusCountsDto.fromMap(taskCountByStatus),
                    TaskPriorityCountsDto.fromMap(taskCountByPriority),
                    overdueCount,
                    avgCompletionTimeHours,
                    memberWorkload == null ? List.of() : memberWorkload);
        }
    }
}

package dev.kylebradshaw.task.repository;

import dev.kylebradshaw.task.entity.Task;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

public interface TaskRepository extends JpaRepository<Task, UUID> {
    List<Task> findByProjectId(UUID projectId);

    void deleteByProjectIdIn(List<UUID> projectIds);

    @Modifying
    @Query("UPDATE Task t SET t.assignee = null WHERE t.assignee.id = :userId")
    void clearAssigneeByUserId(@Param("userId") UUID userId);
}

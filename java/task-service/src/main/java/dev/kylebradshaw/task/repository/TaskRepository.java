package dev.kylebradshaw.task.repository;

import dev.kylebradshaw.task.entity.Task;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface TaskRepository extends JpaRepository<Task, UUID> {
    List<Task> findByProjectId(UUID projectId);
}

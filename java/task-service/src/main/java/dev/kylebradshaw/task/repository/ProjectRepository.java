package dev.kylebradshaw.task.repository;

import dev.kylebradshaw.task.entity.Project;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ProjectRepository extends JpaRepository<Project, UUID> {
}

package dev.kylebradshaw.activity.service;

import dev.kylebradshaw.activity.document.Comment;
import dev.kylebradshaw.activity.repository.CommentRepository;
import java.util.List;
import org.springframework.stereotype.Service;

@Service
public class CommentService {
    private final CommentRepository commentRepo;

    public CommentService(CommentRepository commentRepo) {
        this.commentRepo = commentRepo;
    }

    public Comment addComment(String taskId, String authorId, String body) {
        return commentRepo.save(new Comment(taskId, authorId, body));
    }

    public List<Comment> getCommentsByTask(String taskId) {
        return commentRepo.findByTaskIdOrderByCreatedAtAsc(taskId);
    }
}

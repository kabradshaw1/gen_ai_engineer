package dev.kylebradshaw.activity.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

import dev.kylebradshaw.activity.document.Comment;
import dev.kylebradshaw.activity.repository.CommentRepository;
import java.util.List;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class CommentServiceTest {
    @Mock private CommentRepository commentRepo;
    @InjectMocks private CommentService commentService;

    @Test
    void addComment_savesAndReturns() {
        String taskId = UUID.randomUUID().toString();
        String authorId = UUID.randomUUID().toString();
        Comment comment = new Comment(taskId, authorId, "Looks good!");
        when(commentRepo.save(any(Comment.class))).thenReturn(comment);

        Comment result = commentService.addComment(taskId, authorId, "Looks good!");
        assertThat(result.getBody()).isEqualTo("Looks good!");
    }

    @Test
    void getCommentsByTask_returnsSorted() {
        String taskId = UUID.randomUUID().toString();
        when(commentRepo.findByTaskIdOrderByCreatedAtAsc(taskId))
                .thenReturn(List.of(new Comment(taskId, "u1", "First"), new Comment(taskId, "u2", "Second")));

        List<Comment> result = commentService.getCommentsByTask(taskId);
        assertThat(result).hasSize(2);
    }
}

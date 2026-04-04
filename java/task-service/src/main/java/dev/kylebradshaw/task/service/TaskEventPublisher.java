package dev.kylebradshaw.task.service;

import dev.kylebradshaw.task.config.RabbitConfig;
import dev.kylebradshaw.task.dto.TaskEventMessage;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.stereotype.Component;

@Component
public class TaskEventPublisher {
    private final RabbitTemplate rabbitTemplate;

    public TaskEventPublisher(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    public void publish(String routingKey, TaskEventMessage message) {
        rabbitTemplate.convertAndSend(RabbitConfig.EXCHANGE_NAME, routingKey, message);
    }
}

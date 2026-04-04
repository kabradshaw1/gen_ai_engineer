package dev.kylebradshaw.notification.config;

import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.core.TopicExchange;
import org.springframework.amqp.support.converter.Jackson2JsonMessageConverter;
import org.springframework.amqp.support.converter.MessageConverter;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitConfig {
    public static final String QUEUE_NAME = "notification.queue";
    public static final String EXCHANGE_NAME = "task.events";

    @Bean
    public TopicExchange taskExchange() {
        return new TopicExchange(EXCHANGE_NAME);
    }

    @Bean
    public Queue notificationQueue() {
        return new Queue(QUEUE_NAME, true);
    }

    @Bean
    public Binding bindingTaskCreated(Queue notificationQueue, TopicExchange taskExchange) {
        return BindingBuilder.bind(notificationQueue).to(taskExchange).with("task.created");
    }

    @Bean
    public Binding bindingTaskAssigned(Queue notificationQueue, TopicExchange taskExchange) {
        return BindingBuilder.bind(notificationQueue).to(taskExchange).with("task.assigned");
    }

    @Bean
    public Binding bindingStatusChanged(Queue notificationQueue, TopicExchange taskExchange) {
        return BindingBuilder.bind(notificationQueue).to(taskExchange).with("task.status_changed");
    }

    @Bean
    public Binding bindingCommentAdded(Queue notificationQueue, TopicExchange taskExchange) {
        return BindingBuilder.bind(notificationQueue).to(taskExchange).with("task.comment_added");
    }

    @Bean
    public MessageConverter jsonMessageConverter() {
        return new Jackson2JsonMessageConverter();
    }
}

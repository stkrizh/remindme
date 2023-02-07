FROM rabbitmq:3.11.7-management

ARG delayed_message_plugin_download_url=https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/3.11.1/rabbitmq_delayed_message_exchange-3.11.1.ez
ARG delayed_message_plugin=rabbitmq_delayed_message_exchange-3.11.1.ez

RUN apt update && apt install -y wget
RUN wget -O /plugins/${delayed_message_plugin} ${delayed_message_plugin_download_url} && \
    chown rabbitmq:rabbitmq /plugins/${delayed_message_plugin} && \
    rabbitmq-plugins enable rabbitmq_delayed_message_exchange
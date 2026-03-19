import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../log/log_event.dart';
import '../settings/settings_button.dart';
import 'prompt_command_task.dart';
import 'prompt_controller.dart';

class PromptPage extends StatefulWidget {
  const PromptPage({
    this.sessionId,
    this.threadId,
    super.key,
  }) : assert(
          (sessionId != null && sessionId != '') ||
              (threadId != null && threadId != ''),
          'PromptPage requires either sessionId or threadId.',
        );

  final String? sessionId;
  final String? threadId;

  @override
  State<PromptPage> createState() => _PromptPageState();
}

class _PromptPageState extends State<PromptPage> {
  PromptController? _controller;
  final TextEditingController _textController = TextEditingController();

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _controller ??= PromptController(
      sessionId: widget.sessionId,
      threadId: widget.threadId,
      restClient: AppScope.servicesOf(context).restClient,
      webSocketClient: AppScope.servicesOf(context).webSocketClient,
    )..load();
  }

  @override
  void dispose() {
    _textController.dispose();
    _controller?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = _controller;
    if (controller == null) {
      return const SizedBox.shrink();
    }

    return AnimatedBuilder(
      animation: controller,
      builder: (BuildContext context, Widget? child) {
        return Scaffold(
          appBar: AppBar(
            title: Text(
              widget.threadId != null ? 'Continue Thread' : 'Send Prompt',
            ),
            actions: const <Widget>[SettingsButton()],
          ),
          body: Column(
            children: <Widget>[
              Padding(
                padding: const EdgeInsets.all(16),
                child: Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          widget.threadId != null
                              ? 'Send prompt to current thread'
                              : 'Send prompt to current session',
                          style:
                              Theme.of(context).textTheme.titleMedium?.copyWith(
                                    fontWeight: FontWeight.w700,
                                  ),
                        ),
                        const SizedBox(height: 12),
                        TextField(
                          controller: _textController,
                          minLines: 3,
                          maxLines: 6,
                          decoration: const InputDecoration(
                            hintText:
                                'Describe the next action you want the agent to take',
                            border: OutlineInputBorder(),
                          ),
                        ),
                        const SizedBox(height: 12),
                        Row(
                          children: <Widget>[
                            Expanded(
                              child: Text(
                                widget.threadId != null
                                    ? 'Thread: ${widget.threadId}'
                                    : 'Session: ${widget.sessionId}',
                                style: Theme.of(context).textTheme.bodySmall,
                              ),
                            ),
                            FilledButton.icon(
                              onPressed: controller.isSending
                                  ? null
                                  : () async {
                                      await controller
                                          .sendPrompt(_textController.text);
                                      if (mounted &&
                                          controller.errorMessage == null) {
                                        _textController.clear();
                                      }
                                    },
                              icon: controller.isSending
                                  ? const SizedBox(
                                      width: 16,
                                      height: 16,
                                      child: CircularProgressIndicator(
                                          strokeWidth: 2),
                                    )
                                  : const Icon(Icons.send_rounded),
                              label: Text(controller.isSending
                                  ? 'Sending...'
                                  : 'Send prompt'),
                            ),
                          ],
                        ),
                        if (controller.errorMessage != null) ...<Widget>[
                          const SizedBox(height: 10),
                          Text(
                            controller.errorMessage!,
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.error,
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
              ),
              Expanded(
                child: controller.isLoading && controller.tasks.isEmpty
                    ? const Center(child: CircularProgressIndicator())
                    : controller.tasks.isEmpty
                        ? const Center(
                            child: Text('No prompt tasks yet.'),
                          )
                        : ListView.builder(
                            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                            itemCount: controller.tasks.length,
                            itemBuilder: (BuildContext context, int index) {
                              final task = controller.tasks[index];
                              final commandEvent = _findEvent(
                                controller.events,
                                task.id,
                                LogEventType.command,
                              );
                              final resultEvent = _findEvent(
                                controller.events,
                                task.id,
                                LogEventType.commandResult,
                              );
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 12),
                                child: _PromptTaskCard(
                                  task: task,
                                  commandEvent: commandEvent,
                                  resultEvent: resultEvent,
                                ),
                              );
                            },
                          ),
              ),
            ],
          ),
        );
      },
    );
  }

  LogEvent? _findEvent(
    List<LogEvent> events,
    String commandId,
    LogEventType type,
  ) {
    for (final LogEvent event in events.reversed) {
      if (event.commandId == commandId && event.type == type) {
        return event;
      }
    }
    return null;
  }
}

class _PromptTaskCard extends StatelessWidget {
  const _PromptTaskCard({
    required this.task,
    required this.commandEvent,
    required this.resultEvent,
  });

  final PromptCommandTask task;
  final LogEvent? commandEvent;
  final LogEvent? resultEvent;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final (bg, fg) = switch (task.status) {
      PromptCommandTaskStatus.pending => (
          scheme.secondaryContainer,
          scheme.onSecondaryContainer,
        ),
      PromptCommandTaskStatus.running => (
          const Color(0xFFE0F2FE),
          const Color(0xFF075985),
        ),
      PromptCommandTaskStatus.success => (
          const Color(0xFFD1FAE5),
          const Color(0xFF065F46),
        ),
      PromptCommandTaskStatus.failed => (
          const Color(0xFFFEE2E2),
          const Color(0xFF991B1B),
        ),
    };

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Expanded(
                  child: _PromptPreview(
                    prompt: task.prompt,
                    textStyle: theme.textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                Chip(
                  label: Text(task.status.name),
                  backgroundColor: bg,
                  labelStyle: TextStyle(color: fg, fontWeight: FontWeight.w700),
                  side: BorderSide.none,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              'Command ID: ${task.id}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: scheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 12),
            _FlowRow(
              title: 'Task',
              body:
                  'Created ${_formatTime(task.createdAt)} · Updated ${_formatTime(task.updatedAt)}',
            ),
            if (commandEvent != null)
              _FlowRow(
                title: 'Dispatched',
                body: commandEvent!.content,
              ),
            if (resultEvent != null)
              _FlowRow(
                title: 'Result',
                body: resultEvent!.content,
                emphasized: task.status == PromptCommandTaskStatus.failed,
              )
            else if (task.result.isNotEmpty)
              _FlowRow(
                title: 'Result',
                body: task.result,
                emphasized: task.status == PromptCommandTaskStatus.failed,
              ),
          ],
        ),
      ),
    );
  }

  String _formatTime(DateTime dateTime) {
    final local = dateTime.toLocal();
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    final second = local.second.toString().padLeft(2, '0');
    return '$hour:$minute:$second';
  }
}

class _PromptPreview extends StatelessWidget {
  const _PromptPreview({
    required this.prompt,
    this.textStyle,
  });

  final String prompt;
  final TextStyle? textStyle;

  @override
  Widget build(BuildContext context) {
    final shouldCollapse = prompt.length > 180 || '\n'.allMatches(prompt).length > 4;
    if (!shouldCollapse) {
      return Text(prompt, style: textStyle);
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          prompt,
          maxLines: 3,
          overflow: TextOverflow.ellipsis,
          style: textStyle,
        ),
        const SizedBox(height: 8),
        TextButton(
          onPressed: () {
            showModalBottomSheet<void>(
              context: context,
              isScrollControlled: true,
              builder: (BuildContext context) {
                return DraggableScrollableSheet(
                  expand: false,
                  initialChildSize: 0.65,
                  maxChildSize: 0.9,
                  minChildSize: 0.45,
                  builder: (BuildContext context, ScrollController controller) {
                    return Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Text(
                            'Full Prompt',
                            style: Theme.of(context).textTheme.titleMedium,
                          ),
                          const SizedBox(height: 12),
                          Expanded(
                            child: SingleChildScrollView(
                              controller: controller,
                              child: SelectableText(prompt),
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                );
              },
            );
          },
          child: const Text('View Full Prompt'),
        ),
      ],
    );
  }
}

class _FlowRow extends StatelessWidget {
  const _FlowRow({
    required this.title,
    required this.body,
    this.emphasized = false,
  });

  final String title;
  final String body;
  final bool emphasized;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            width: 72,
            child: Text(
              title,
              style: theme.textTheme.labelMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          Expanded(
            child: Text(
              body,
              style: theme.textTheme.bodyMedium?.copyWith(
                color: emphasized ? theme.colorScheme.error : null,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

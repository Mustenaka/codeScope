import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../thread/thread_record.dart';
import 'thread_detail_controller.dart';
import 'thread_message_record.dart';

class ThreadDetailPage extends StatefulWidget {
  const ThreadDetailPage({
    required this.thread,
    super.key,
  });

  final ThreadRecord thread;

  @override
  State<ThreadDetailPage> createState() => _ThreadDetailPageState();
}

class _ThreadDetailPageState extends State<ThreadDetailPage> {
  ThreadDetailController? _controller;
  final ScrollController _scrollController = ScrollController();

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _controller ??= ThreadDetailController(
      AppScope.servicesOf(context).restClient,
      AppScope.servicesOf(context).webSocketClient,
    )..load(widget.thread.id);
  }

  @override
  void dispose() {
    _controller?.dispose();
    _scrollController.dispose();
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
        final thread = controller.thread ?? widget.thread;
        return Scaffold(
          appBar: AppBar(title: Text(thread.title)),
          body: Column(
            children: <Widget>[
              ListTile(
                title: Text(thread.status.label),
                subtitle: Text(thread.displaySummary),
              ),
              Expanded(
                child: controller.isLoading && controller.messages.isEmpty
                    ? const Center(child: CircularProgressIndicator())
                    : controller.errorMessage != null &&
                            controller.messages.isEmpty
                        ? Center(child: Text(controller.errorMessage!))
                        : controller.messages.isEmpty
                            ? const _EmptyMessageState()
                            : Scrollbar(
                                controller: _scrollController,
                                thumbVisibility: true,
                                child: ListView.builder(
                                  controller: _scrollController,
                                  padding:
                                      const EdgeInsets.fromLTRB(16, 8, 16, 24),
                                  itemCount: controller.messages.length,
                                  itemBuilder:
                                      (BuildContext context, int index) {
                                    final message = controller.messages[index];
                                    return _MessageBubble(
                                      key: ValueKey<String>(message.id),
                                      message: message,
                                    );
                                  },
                                ),
                              ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _EmptyMessageState extends StatelessWidget {
  const _EmptyMessageState();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(
              Icons.chat_bubble_outline,
              size: 40,
              color: theme.colorScheme.outline,
            ),
            const SizedBox(height: 12),
            Text(
              'No readable message history yet.',
              style: theme.textTheme.titleMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'This thread is still using partial bridge capture, so real user and assistant messages may not be available yet.',
              style: theme.textTheme.bodyMedium,
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

class _MessageBubble extends StatefulWidget {
  const _MessageBubble({
    required this.message,
    super.key,
  });

  final ThreadMessageRecord message;

  @override
  State<_MessageBubble> createState() => _MessageBubbleState();
}

class _MessageBubbleState extends State<_MessageBubble> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final message = widget.message;
    final isUser = message.role == ThreadMessageRole.user;
    final shouldCollapse = _shouldCollapse(message.content);
    final availableWidth = MediaQuery.sizeOf(context).width * 0.82;

    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
        constraints: BoxConstraints(maxWidth: availableWidth.clamp(260, 560)),
        decoration: BoxDecoration(
          color: isUser
              ? theme.colorScheme.primaryContainer
              : theme.colorScheme.surfaceContainerHighest,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          crossAxisAlignment:
              isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: <Widget>[
            Text(message.role.label, style: theme.textTheme.labelMedium),
            const SizedBox(height: 6),
            _expanded || !shouldCollapse
                ? SelectableText(message.content)
                : Text(
                    message.content,
                    maxLines: 12,
                    overflow: TextOverflow.fade,
                  ),
            if (shouldCollapse) ...<Widget>[
              const SizedBox(height: 10),
              TextButton(
                onPressed: () {
                  setState(() {
                    _expanded = !_expanded;
                  });
                },
                child: Text(_expanded ? 'Collapse' : 'Expand'),
              ),
            ],
          ],
        ),
      ),
    );
  }

  bool _shouldCollapse(String content) {
    const maxLength = 800;
    const maxLineBreaks = 12;
    return content.length > maxLength ||
        '\n'.allMatches(content).length > maxLineBreaks;
  }
}

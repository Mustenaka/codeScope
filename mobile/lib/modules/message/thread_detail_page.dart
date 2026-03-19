import 'package:flutter/material.dart';

import '../../app/app_scope.dart';
import '../../app/display_text.dart';
import '../prompt/prompt_page.dart';
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
    )..load(widget.thread.id, initialThread: widget.thread);
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
          appBar: AppBar(
            title: Text(thread.title),
            actions: <Widget>[
              IconButton(
                tooltip: 'Continue with prompt',
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute<void>(
                      builder: (_) => PromptPage(threadId: thread.id),
                    ),
                  );
                },
                icon: const Icon(Icons.bolt_rounded),
              ),
            ],
          ),
          body: Column(
            children: <Widget>[
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: <Widget>[
                            _ThreadMetaChip(label: thread.status.label),
                            _ThreadMetaChip(label: thread.agentLabel),
                            _ThreadMetaChip(
                              label:
                                  'Started ${formatTimestamp(thread.startedAt)}',
                            ),
                            _ThreadMetaChip(
                              label:
                                  'Active ${formatTimestamp(thread.lastActivityAt)}',
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        _ThreadSummaryPreview(summary: thread.displaySummary),
                      ],
                    ),
                  ),
                ),
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
            if (message.role == ThreadMessageRole.assistant) ...<Widget>[
              const SizedBox(height: 4),
              _AssistantBadge(label: message.sourceLabel),
            ],
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
              Wrap(
                spacing: 8,
                children: <Widget>[
                  TextButton(
                    onPressed: () {
                      setState(() {
                        _expanded = !_expanded;
                      });
                    },
                    child: Text(_expanded ? 'Collapse' : 'Expand'),
                  ),
                  if (!_expanded)
                    TextButton(
                      onPressed: () => _showFullMessageSheet(context, message),
                      child: const Text('View Full Message'),
                    ),
                ],
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

  void _showFullMessageSheet(BuildContext context, ThreadMessageRecord message) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (BuildContext context) {
        return DraggableScrollableSheet(
          expand: false,
          initialChildSize: 0.72,
          maxChildSize: 0.92,
          minChildSize: 0.45,
          builder: (BuildContext context, ScrollController controller) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    'Full Message',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: <Widget>[
                      _ThreadMetaChip(label: message.role.label),
                      if (message.role == ThreadMessageRole.assistant)
                        _ThreadMetaChip(label: message.sourceLabel),
                      _ThreadMetaChip(
                        label: formatTimestamp(message.createdAt),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Expanded(
                    child: SingleChildScrollView(
                      controller: controller,
                      child: SelectableText(message.content),
                    ),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }
}

class _ThreadMetaChip extends StatelessWidget {
  const _ThreadMetaChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(label, style: Theme.of(context).textTheme.bodySmall),
    );
  }
}

class _AssistantBadge extends StatelessWidget {
  const _AssistantBadge({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.tertiaryContainer,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
      ),
    );
  }
}

class _ThreadSummaryPreview extends StatelessWidget {
  const _ThreadSummaryPreview({required this.summary});

  final String summary;

  @override
  Widget build(BuildContext context) {
    final shouldCollapse =
        summary.length > 180 || '\n'.allMatches(summary).length > 3;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          summary,
          maxLines: shouldCollapse ? 3 : null,
          overflow: shouldCollapse ? TextOverflow.ellipsis : TextOverflow.visible,
        ),
        if (shouldCollapse) ...<Widget>[
          const SizedBox(height: 8),
          TextButton(
            onPressed: () {
              showModalBottomSheet<void>(
                context: context,
                isScrollControlled: true,
                builder: (BuildContext context) {
                  return DraggableScrollableSheet(
                    expand: false,
                    initialChildSize: 0.5,
                    maxChildSize: 0.85,
                    minChildSize: 0.35,
                    builder:
                        (BuildContext context, ScrollController controller) {
                      return Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            Text(
                              'Thread Details',
                              style: Theme.of(context).textTheme.titleMedium,
                            ),
                            const SizedBox(height: 12),
                            Expanded(
                              child: SingleChildScrollView(
                                controller: controller,
                                child: SelectableText(summary),
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
            child: const Text('View Details'),
          ),
        ],
      ],
    );
  }
}

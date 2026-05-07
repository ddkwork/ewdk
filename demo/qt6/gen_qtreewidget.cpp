#include <QTreeWidget>
#include <QTreeWidgetItem>
#include <QHeaderView>
#include <QMenu>
#include <QAction>
#include <QString>
#include <QKeySequence>
#include <QPoint>
#include <QCursor>
#include <QMap>
#include <QList>
#include "gen_qtreewidget.h"
#include "libmiqt.h"

#ifdef __cplusplus
extern "C" {
#endif

QTreeWidgetItem* QTreeWidgetItem_new() {
	return new QTreeWidgetItem();
}

QTreeWidgetItem* QTreeWidgetItem_new2(QTreeWidget* view) {
	return new QTreeWidgetItem(view);
}

QTreeWidgetItem* QTreeWidgetItem_new3(QTreeWidget* view, struct miqt_string text) {
	return new QTreeWidgetItem(view, QStringList(QString::fromUtf8(text.data, (int)text.len)));
}

QTreeWidgetItem* QTreeWidgetItem_new4(QTreeWidgetItem* parent) {
	return new QTreeWidgetItem(parent);
}

QTreeWidgetItem* QTreeWidgetItem_new5(QTreeWidgetItem* parent, struct miqt_string text) {
	return new QTreeWidgetItem(parent, QStringList(QString::fromUtf8(text.data, (int)text.len)));
}

void QTreeWidgetItem_setText(QTreeWidgetItem* self, int column, struct miqt_string text) {
	self->setText(column, QString::fromUtf8(text.data, (int)text.len));
}

void QWidgetItem_setTextAlignment(QTreeWidgetItem* self, int column, int alignment) {
	self->setTextAlignment(column, (Qt::AlignmentFlag)alignment);
}

void QWidgetItem_setIcon(QTreeWidgetItem* self, int column, int iconId) {
	Q_UNUSED(iconId);
	self->setIcon(column, QIcon());
}

void QWidgetItem_setToolTip(QTreeWidgetItem* self, int column, struct miqt_string tip) {
	self->setToolTip(column, QString::fromUtf8(tip.data, (int)tip.len));
}

void QWidgetItem_setFlags(QTreeWidgetItem* self, int flags) {
	self->setFlags((Qt::ItemFlags)flags);
}

void QWidgetItem_setCheckState(QTreeWidgetItem* self, int column, int state) {
	self->setCheckState(column, (Qt::CheckState)state);
}

int QWidgetItem_checkState(QTreeWidgetItem* self, int column) {
	return (int)self->checkState(column);
}

void QWidgetItem_setData(QTreeWidgetItem* self, int column, int role, int value) {
	self->setData(column, (Qt::ItemDataRole)role, value);
}

int QWidgetItem_data(QTreeWidgetItem* self, int column, int role) {
	return self->data(column, (Qt::ItemDataRole)role).toInt();
}

void QWidgetItem_addChild(QTreeWidgetItem* self, QTreeWidgetItem* child) {
	self->addChild(child);
}

void QWidgetItem_insertChild(QTreeWidgetItem* self, int index, QTreeWidgetItem* child) {
	self->insertChild(index, child);
}

void QWidgetItem_removeChild(QTreeWidgetItem* self, QTreeWidgetItem* child) {
	self->removeChild(child);
}

int QWidgetItem_childCount(QTreeWidgetItem* self) {
	return self->childCount();
}

QTreeWidgetItem* QWidgetItem_child(QTreeWidgetItem* self, int index) {
	return self->child(index);
}

int QWidgetItem_columnCount(QTreeWidgetItem* self) {
	return self->columnCount();
}

void QWidgetItem_sortChildren(QTreeWidgetItem* self, int column, int order) {
	self->sortChildren(column, (Qt::SortOrder)order);
}

QTreeWidget* QTreeWidget_new() {
	return new QTreeWidget();
}

QTreeWidget* QTreeWidget_new2(QWidget* parent) {
	return new QTreeWidget(parent);
}

void QTreeWidget_setHeaderLabels(QTreeWidget* self, struct miqt_array labels) {
	QStringList list;
	if (labels.len > 0 && labels.data != nullptr) {
		miqt_string* arr = static_cast<miqt_string*>(labels.data);
		for (size_t i = 0; i < labels.len; i++) {
			list.append(QString::fromUtf8(arr[i].data, (int)arr[i].len));
		}
	}
	self->setHeaderLabels(list);
}

void QTreeWidget_setColumnWidth(QTreeWidget* self, int column, int width) {
	self->setColumnWidth(column, width);
}

void QTreeWidget_addTopLevelItem(QTreeWidget* self, QTreeWidgetItem* item) {
	self->addTopLevelItem(item);
}

void QTreeWidget_insertTopLevelItem(QTreeWidget* self, int index, QTreeWidgetItem* item) {
	self->insertTopLevelItem(index, item);
}

void QTreeWidget_takeTopLevelItem(QTreeWidget* self, int index) {
	self->takeTopLevelItem(index);
}

int QTreeWidget_topLevelItemCount(QTreeWidget* self) {
	return self->topLevelItemCount();
}

QTreeWidgetItem* QTreeWidget_topLevelItem(QTreeWidget* self, int index) {
	return self->topLevelItem(index);
}

QTreeWidgetItem* QTreeWidget_currentItem(QTreeWidget* self) {
	return self->currentItem();
}

void QTreeWidget_setCurrentItem(QTreeWidget* self, QTreeWidgetItem* item) {
	self->setCurrentItem(item);
}

void QTreeWidget_expandAll(QTreeWidget* self) {
	self->expandAll();
}

void QTreeWidget_collapseAll(QTreeWidget* self) {
	self->collapseAll();
}

void QTreeWidget_expand(QTreeWidget* self, QTreeWidgetItem* item) {
	self->expandItem(item);
}

void QTreeWidget_collapse(QTreeWidget* self, QTreeWidgetItem* item) {
	self->collapseItem(item);
}

int QTreeWidget_isExpanded(QTreeWidget* self, QTreeWidgetItem* item) {
	return item->isExpanded() ? 1 : 0;
}

void QTreeWidget_sortByColumn(QTreeWidget* self, int column, int order) {
	self->sortByColumn(column, (Qt::SortOrder)order);
}

void QTreeWidget_sortItems(QTreeWidget* self, int column, int order) {
	self->sortItems(column, (Qt::SortOrder)order);
}

void QTreeWidget_setSortingEnabled(QTreeWidget* self, int enabled) {
	self->setSortingEnabled(enabled != 0);
}

int QTreeWidget_isSortingEnabled(QTreeWidget* self) {
	return self->isSortingEnabled() ? 1 : 0;
}

void QTreeWidget_setAlternatingRowColors(QTreeWidget* self, int enable) {
	self->setAlternatingRowColors(enable != 0);
}

void QTreeWidget_setAllColumnsShowFocus(QTreeWidget* self, int show) {
	self->setAllColumnsShowFocus(show != 0);
}

void QTreeWidget_setRootIsDecorated(QTreeWidget* self, int show) {
	self->setRootIsDecorated(show != 0);
}

void QTreeWidget_setItemsExpandable(QTreeWidget* self, int enable) {
	self->setItemsExpandable(enable != 0);
}

void QTreeWidget_setAnimated(QTreeWidget* self, int animate) {
	self->setAnimated(animate != 0);
}

void QTreeWidget_setSelectionBehavior(QTreeWidget* self, int behavior) {
	self->setSelectionBehavior((QAbstractItemView::SelectionBehavior)behavior);
}

void QTreeWidget_setSelectionMode(QTreeWidget* self, int mode) {
	self->setSelectionMode((QAbstractItemView::SelectionMode)mode);
}

void QTreeWidget_setEditTriggers(QTreeWidget* self, int triggers) {
	self->setEditTriggers((QAbstractItemView::EditTriggers)triggers);
}

void QTreeWidget_setWordWrap(QTreeWidget* self, int on) {
	self->setWordWrap(on != 0);
}

void QTreeWidget_resizeColumnToContents(QTreeWidget* self, int column) {
	self->resizeColumnToContents(column);
}

void QTreeWidget_headerResizeSections(QTreeWidget* self) {
	self->header()->resizeSections(QHeaderView::ResizeToContents);
}

void QTreeWidget_headerSetSectionResizeMode(QTreeWidget* self, int logicalIndex, int mode) {
	self->header()->setSectionResizeMode(logicalIndex, (QHeaderView::ResizeMode)mode);
}

QHeaderView* QTreeWidget_header(QTreeWidget* self) {
	return self->header();
}

void QTreeWidget_setHeaderHidden(QTreeWidget* self, int hide) {
	self->setHeaderHidden(hide != 0);
}

void QTreeWidget_setContextMenuPolicy2(QTreeWidget* self, int policy) {
	self->setContextMenuPolicy((Qt::ContextMenuPolicy)policy);
}

void QTreeWidget_clear(QTreeWidget* self) {
	self->clear();
}

void QTreeWidget_setColumnCount(QTreeWidget* self, int columns) {
	self->setColumnCount(columns);
}

int QTreeWidget_columnCount(QTreeWidget* self) {
	return self->columnCount();
}

void QTreeWidget_setFirstColumnSpanned(QTreeWidget* self, int row, QTreeWidgetItem* parent, int spanned) {
	self->setFirstColumnSpanned(row, self->indexFromItem(parent), spanned != 0);
}

QMenu* QMenu_new(QWidget* parent) {
	return new QMenu(parent);
}

void QMenu_addAction(QMenu* self, QAction* action) {
	self->addAction(action);
}

void QMenu_addSeparator(QMenu* self) {
	self->addSeparator();
}

QAction* QMenu_exec(QMenu* self, int x, int y) {
	return (QAction*)self->exec(QPoint(x, y));
}

void QMenu_exec2(QMenu* self, int x, int y, QAction* action) {
	self->exec(QPoint(x, y), action);
}

QAction* QAction_new(struct miqt_string text, QObject* parent) {
	return new QAction(QString::fromUtf8(text.data, (int)text.len), parent);
}

QAction* QAction_new2(QObject* parent) {
	return new QAction(parent);
}

void QAction_setText(QAction* self, struct miqt_string text) {
	self->setText(QString::fromUtf8(text.data, (int)text.len));
}

void QAction_setCheckable(QAction* self, int checkable) {
	self->setCheckable(checkable != 0);
}

void QAction_setChecked(QAction* self, int checked) {
	self->setChecked(checked != 0);
}

void QAction_setEnabled(QAction* self, int enabled) {
	self->setEnabled(enabled != 0);
}

void QAction_setShortcut(QAction* self, struct miqt_string key) {
	self->setShortcut(QKeySequence(QString::fromUtf8(key.data, (int)key.len)));
}

int QAction_isChecked(QAction* self) {
	return self->isChecked() ? 1 : 0;
}

int QAction_isEnabled(QAction* self) {
	return self->isEnabled() ? 1 : 0;
}

void QAction_trigger(QAction* self) {
	self->trigger();
}

void QTreeWidget_installContextMenu(QTreeWidget* self) {
	QObject::connect(self, &QTreeWidget::customContextMenuRequested, [self](const QPoint& pos) {
		QMenu* menu = new QMenu(self);
		menu->addAction(new QAction(QStringLiteral("Expand All"), menu));
		menu->addAction(new QAction(QStringLiteral("Collapse All"), menu));
		menu->addSeparator();
		menu->addAction(new QAction(QStringLiteral("Sort Ascending"), menu));
		menu->addAction(new QAction(QStringLiteral("Sort Descending"), menu));
		menu->addSeparator();
		QAction* copyAction = new QAction(QStringLiteral("Copy Path (Ctrl+C)"), menu);
		copyAction->setShortcut(QKeySequence(Qt::CTRL | Qt::Key_C));
		menu->addAction(copyAction);
		menu->addSeparator();
		QAction* refreshAction = new QAction(QStringLiteral("Refresh (F5)"), menu);
		refreshAction->setShortcut(QKeySequence(Qt::Key_F5));
		menu->addAction(refreshAction);

		QAction* selected = menu->exec(self->viewport()->mapToGlobal(pos));
		if (selected) {
			QString text = selected->text();
			if (text == QStringLiteral("Expand All")) {
				self->expandAll();
			} else if (text == QStringLiteral("Collapse All")) {
				self->collapseAll();
			} else if (text == QStringLiteral("Sort Ascending")) {
				int col = self->header()->sortIndicatorSection();
				self->sortByColumn(col, Qt::AscendingOrder);
			} else if (text == QStringLiteral("Sort Descending")) {
				int col = self->header()->sortIndicatorSection();
				self->sortByColumn(col, Qt::DescendingOrder);
			}
		}
		delete menu;
	});
}

void QTreeWidget_showGridLines(QTreeWidget* self) {
	self->setStyleSheet(
		QLatin1String("QTreeView::item {"
		"  border-right: 0.5px solid #cccccc;"
		"}")
	);
}

void QTreeWidget_enableDragDrop(QTreeWidget* self) {
	self->setDragEnabled(true);
	self->setAcceptDrops(true);
	self->setDropIndicatorShown(true);
	self->setDragDropMode(QAbstractItemView::InternalMove);
	self->setSelectionBehavior(QAbstractItemView::SelectRows);
}

struct ContextMenuItem {
	QString text;
	bool isSeparator;
	int shortcutKey;
	int shortcutMod;
};

static QMap<QTreeWidget*, QList<ContextMenuItem>> g_contextMenus;
static QMap<QTreeWidget*, intptr_t> g_contextCallbacks;

void QTreeWidget_addContextMenuItem(QTreeWidget* self, struct miqt_string text, int isSeparator, int shortcutKey, int shortcutMod) {
	ContextMenuItem item;
	item.text = QString::fromUtf8(text.data, text.len);
	item.isSeparator = (isSeparator != 0);
	item.shortcutKey = shortcutKey;
	item.shortcutMod = shortcutMod;
	g_contextMenus[self].append(item);
}

void QTreeWidget_clearContextMenuItems(QTreeWidget* self) {
	g_contextMenus.remove(self);
	g_contextCallbacks.remove(self);
}

void QTreeWidget_installContextMenuWithCallback(QTreeWidget* self, int64_t cb) {
	g_contextCallbacks[self] = (intptr_t)cb;
	QObject::connect(self, &QTreeWidget::customContextMenuRequested, [self](const QPoint& pos) {
		QMenu* menu = new QMenu(self);
		const QList<ContextMenuItem>& items = g_contextMenus.value(self);
		for (int i = 0; i < items.size(); ++i) {
			const ContextMenuItem& item = items[i];
			if (item.isSeparator) {
				menu->addSeparator();
			} else {
				QAction* action = new QAction(item.text, menu);
				if (item.shortcutKey > 0 || item.shortcutMod > 0) {
					action->setShortcut(QKeySequence((Qt::Key)item.shortcutKey, (Qt::KeyboardModifiers)item.shortcutMod));
				}
				action->setData(i);
				menu->addAction(action);
			}
		}
		QAction* selected = menu->exec(self->viewport()->mapToGlobal(pos));
		if (selected && g_contextCallbacks.contains(self)) {
			int idx = selected->data().toInt();
			typedef void(*ContextMenuCallback)(int);
			((ContextMenuCallback)g_contextCallbacks[self])(idx);
		}
		delete menu;
	});
}

#ifdef __cplusplus
}
#endif

#pragma once
#ifndef MIQT_QT6_GEN_QTREEWIDGET_H
#define MIQT_QT6_GEN_QTREEWIDGET_H

#include "libmiqt.h"
#include "miqt_export.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
class QMenu;
class QAction;
class QHeaderView;
class QTreeWidget;
class QTreeWidgetItem;
class QVariant;
#else
typedef struct QMenu QMenu;
typedef struct QAction QAction;
typedef struct QHeaderView QHeaderView;
typedef struct QTreeWidget QTreeWidget;
typedef struct QTreeWidgetItem QTreeWidgetItem;
typedef struct QVariant QVariant;
#endif

MIQT_EXPORT QTreeWidgetItem*QTreeWidgetItem_new();
MIQT_EXPORT QTreeWidgetItem*QTreeWidgetItem_new2(QTreeWidget* view);
MIQT_EXPORT QTreeWidgetItem*QTreeWidgetItem_new3(QTreeWidget* view, struct miqt_string text);
MIQT_EXPORT QTreeWidgetItem*QTreeWidgetItem_new4(QTreeWidgetItem* parent);
MIQT_EXPORT QTreeWidgetItem*QTreeWidgetItem_new5(QTreeWidgetItem* parent, struct miqt_string text);
MIQT_EXPORT void QTreeWidgetItem_setText(QTreeWidgetItem* self, int column, struct miqt_string text);
MIQT_EXPORT void QWidgetItem_setTextAlignment(QTreeWidgetItem* self, int column, int alignment);
MIQT_EXPORT void QWidgetItem_setIcon(QTreeWidgetItem* self, int column, int iconId);
MIQT_EXPORT void QWidgetItem_setToolTip(QTreeWidgetItem* self, int column, struct miqt_string tip);
MIQT_EXPORT void QWidgetItem_setFlags(QTreeWidgetItem* self, int flags);
MIQT_EXPORT void QWidgetItem_setCheckState(QTreeWidgetItem* self, int column, int state);
MIQT_EXPORT int QWidgetItem_checkState(QTreeWidgetItem* self, int column);
MIQT_EXPORT void QWidgetItem_setData(QTreeWidgetItem* self, int column, int role, int value);
MIQT_EXPORT int QWidgetItem_data(QTreeWidgetItem* self, int column, int role);
MIQT_EXPORT void QWidgetItem_addChild(QTreeWidgetItem* self, QTreeWidgetItem* child);
MIQT_EXPORT void QWidgetItem_insertChild(QTreeWidgetItem* self, int index, QTreeWidgetItem* child);
MIQT_EXPORT void QWidgetItem_removeChild(QTreeWidgetItem* self, QTreeWidgetItem* child);
MIQT_EXPORT int QWidgetItem_childCount(QTreeWidgetItem* self);
MIQT_EXPORT QTreeWidgetItem*QWidgetItem_child(QTreeWidgetItem* self, int index);
MIQT_EXPORT int QWidgetItem_columnCount(QTreeWidgetItem* self);
MIQT_EXPORT void QWidgetItem_sortChildren(QTreeWidgetItem* self, int column, int order);

MIQT_EXPORT QTreeWidget*QTreeWidget_new();
MIQT_EXPORT QTreeWidget*QTreeWidget_new2(QWidget* parent);
MIQT_EXPORT void QTreeWidget_setHeaderLabels(QTreeWidget* self, struct miqt_array labels);
MIQT_EXPORT void QTreeWidget_setColumnWidth(QTreeWidget* self, int column, int width);
MIQT_EXPORT void QTreeWidget_addTopLevelItem(QTreeWidget* self, QTreeWidgetItem* item);
MIQT_EXPORT void QTreeWidget_insertTopLevelItem(QTreeWidget* self, int index, QTreeWidgetItem* item);
MIQT_EXPORT void QTreeWidget_takeTopLevelItem(QTreeWidget* self, int index);
MIQT_EXPORT int QTreeWidget_topLevelItemCount(QTreeWidget* self);
MIQT_EXPORT QTreeWidgetItem*QTreeWidget_topLevelItem(QTreeWidget* self, int index);
MIQT_EXPORT QTreeWidgetItem*QTreeWidget_currentItem(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_setCurrentItem(QTreeWidget* self, QTreeWidgetItem* item);
MIQT_EXPORT void QTreeWidget_expandAll(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_collapseAll(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_expand(QTreeWidget* self, QTreeWidgetItem* item);
MIQT_EXPORT void QTreeWidget_collapse(QTreeWidget* self, QTreeWidgetItem* item);
MIQT_EXPORT int QTreeWidget_isExpanded(QTreeWidget* self, QTreeWidgetItem* item);
MIQT_EXPORT void QTreeWidget_sortByColumn(QTreeWidget* self, int column, int order);
MIQT_EXPORT void QTreeWidget_sortItems(QTreeWidget* self, int column, int order);
MIQT_EXPORT void QTreeWidget_setSortingEnabled(QTreeWidget* self, int enabled);
MIQT_EXPORT int QTreeWidget_isSortingEnabled(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_setAlternatingRowColors(QTreeWidget* self, int enable);
MIQT_EXPORT void QTreeWidget_setAllColumnsShowFocus(QTreeWidget* self, int show);
MIQT_EXPORT void QTreeWidget_setRootIsDecorated(QTreeWidget* self, int show);
MIQT_EXPORT void QTreeWidget_setItemsExpandable(QTreeWidget* self, int enable);
MIQT_EXPORT void QTreeWidget_setAnimated(QTreeWidget* self, int animate);
MIQT_EXPORT void QTreeWidget_setSelectionBehavior(QTreeWidget* self, int behavior);
MIQT_EXPORT void QTreeWidget_setSelectionMode(QTreeWidget* self, int mode);
MIQT_EXPORT void QTreeWidget_setEditTriggers(QTreeWidget* self, int triggers);
MIQT_EXPORT void QTreeWidget_setWordWrap(QTreeWidget* self, int on);
MIQT_EXPORT void QTreeWidget_resizeColumnToContents(QTreeWidget* self, int column);
MIQT_EXPORT void QTreeWidget_headerResizeSections(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_headerSetSectionResizeMode(QTreeWidget* self, int logicalIndex, int mode);
MIQT_EXPORT QHeaderView*QTreeWidget_header(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_setHeaderHidden(QTreeWidget* self, int hide);
MIQT_EXPORT void QTreeWidget_setContextMenuPolicy2(QTreeWidget* self, int policy);
MIQT_EXPORT void QTreeWidget_clear(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_setColumnCount(QTreeWidget* self, int columns);
MIQT_EXPORT int QTreeWidget_columnCount(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_setFirstColumnSpanned(QTreeWidget* self, int row, QTreeWidgetItem* parent, int spanned);
MIQT_EXPORT void QTreeWidget_installContextMenu(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_showGridLines(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_enableDragDrop(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_addContextMenuItem(QTreeWidget* self, struct miqt_string text, int isSeparator, int shortcutKey, int shortcutMod);
MIQT_EXPORT void QTreeWidget_clearContextMenuItems(QTreeWidget* self);
MIQT_EXPORT void QTreeWidget_installContextMenuWithCallback(QTreeWidget* self, int64_t cb);

MIQT_EXPORT QMenu*QMenu_new(QWidget* parent);
MIQT_EXPORT void QMenu_addAction(QMenu* self, QAction* action);
MIQT_EXPORT void QMenu_addSeparator(QMenu* self);
MIQT_EXPORT QAction*QMenu_exec(QMenu* self, int x, int y);
MIQT_EXPORT void QMenu_exec2(QMenu* self, int x, int y, QAction* action);

MIQT_EXPORT QAction*QAction_new(struct miqt_string text, QObject* parent);
MIQT_EXPORT QAction*QAction_new2(QObject* parent);
MIQT_EXPORT void QAction_setText(QAction* self, struct miqt_string text);
MIQT_EXPORT void QAction_setCheckable(QAction* self, int checkable);
MIQT_EXPORT void QAction_setChecked(QAction* self, int checked);
MIQT_EXPORT void QAction_setEnabled(QAction* self, int enabled);
MIQT_EXPORT void QAction_setShortcut(QAction* self, struct miqt_string key);
MIQT_EXPORT int QAction_isChecked(QAction* self);
MIQT_EXPORT int QAction_isEnabled(QAction* self);
MIQT_EXPORT void QAction_trigger(QAction* self);

#ifdef __cplusplus
}
#endif

#endif

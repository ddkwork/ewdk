#pragma once
#ifndef MIQT_QT6_GEN_QPUSHBUTTON_H
#define MIQT_QT6_GEN_QPUSHBUTTON_H




#include "libmiqt.h"
#include "miqt_export.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
class QEvent;
class QFocusEvent;
class QIcon;
class QKeyEvent;
class QMenu;
class QMetaObject;
class QMouseEvent;
class QPaintEvent;
class QPoint;
class QPushButton;
class QSize;
class QStyleOptionButton;
class QWidget;
#else
typedef struct QEvent QEvent;
typedef struct QFocusEvent QFocusEvent;
typedef struct QIcon QIcon;
typedef struct QKeyEvent QKeyEvent;
typedef struct QMenu QMenu;
typedef struct QMetaObject QMetaObject;
typedef struct QMouseEvent QMouseEvent;
typedef struct QPaintEvent QPaintEvent;
typedef struct QPoint QPoint;
typedef struct QPushButton QPushButton;
typedef struct QSize QSize;
typedef struct QStyleOptionButton QStyleOptionButton;
typedef struct QWidget QWidget;
#endif

MIQT_EXPORT QPushButton*QPushButton_new(QWidget*parent);
MIQT_EXPORT QPushButton*QPushButton_new2();
MIQT_EXPORT QPushButton*QPushButton_new3(struct miqt_string text);
MIQT_EXPORT QPushButton*QPushButton_new4(QIcon*icon, struct miqt_string text);
MIQT_EXPORT QPushButton*QPushButton_new5(struct miqt_string text, QWidget*parent);
MIQT_EXPORT QPushButton*QPushButton_new6(QIcon*icon, struct miqt_string text, QWidget*parent);
MIQT_EXPORT QMetaObject*QPushButton_metaObject(QPushButton*self);
MIQT_EXPORT void*QPushButton_metacast(QPushButton*self, char*param1);
MIQT_EXPORT struct miqt_string QPushButton_tr(char*s);
MIQT_EXPORT QSize*QPushButton_sizeHint(QPushButton*self);
MIQT_EXPORT QSize*QPushButton_minimumSizeHint(QPushButton*self);
MIQT_EXPORT bool QPushButton_autoDefault(QPushButton*self);
MIQT_EXPORT void QPushButton_setAutoDefault(QPushButton*self, bool autoDefault);
MIQT_EXPORT bool QPushButton_isDefault(QPushButton*self);
MIQT_EXPORT void QPushButton_setDefault(QPushButton*self, bool defaultVal);
MIQT_EXPORT void QPushButton_setMenu(QPushButton*self, QMenu*menu);
MIQT_EXPORT QMenu*QPushButton_menu(QPushButton*self);
MIQT_EXPORT void QPushButton_setFlat(QPushButton*self, bool flat);
MIQT_EXPORT bool QPushButton_isFlat(QPushButton*self);
MIQT_EXPORT void QPushButton_showMenu(QPushButton*self);
MIQT_EXPORT bool QPushButton_event(QPushButton*self, QEvent*e);
MIQT_EXPORT void QPushButton_paintEvent(QPushButton*self, QPaintEvent*param1);
MIQT_EXPORT void QPushButton_keyPressEvent(QPushButton*self, QKeyEvent*param1);
MIQT_EXPORT void QPushButton_focusInEvent(QPushButton*self, QFocusEvent*param1);
MIQT_EXPORT void QPushButton_focusOutEvent(QPushButton*self, QFocusEvent*param1);
MIQT_EXPORT void QPushButton_mouseMoveEvent(QPushButton*self, QMouseEvent*param1);
MIQT_EXPORT void QPushButton_initStyleOption(QPushButton*self, QStyleOptionButton*option);
MIQT_EXPORT bool QPushButton_hitButton(QPushButton*self, QPoint*pos);
MIQT_EXPORT struct miqt_string QPushButton_tr2(char*s, char*c);
MIQT_EXPORT struct miqt_string QPushButton_tr3(char*s, char*c, int n);

MIQT_EXPORT bool QPushButton_override_virtual_sizeHint(void*self, intptr_t slot);
MIQT_EXPORT QSize*QPushButton_virtualbase_sizeHint(void*self);
MIQT_EXPORT bool QPushButton_override_virtual_minimumSizeHint(void*self, intptr_t slot);
MIQT_EXPORT QSize*QPushButton_virtualbase_minimumSizeHint(void*self);
MIQT_EXPORT bool QPushButton_override_virtual_event(void*self, intptr_t slot);
MIQT_EXPORT bool QPushButton_virtualbase_event(void*self, QEvent*e);
MIQT_EXPORT bool QPushButton_override_virtual_paintEvent(void*self, intptr_t slot);
MIQT_EXPORT void QPushButton_virtualbase_paintEvent(void*self, QPaintEvent*param1);
MIQT_EXPORT bool QPushButton_override_virtual_keyPressEvent(void*self, intptr_t slot);
MIQT_EXPORT void QPushButton_virtualbase_keyPressEvent(void*self, QKeyEvent*param1);
MIQT_EXPORT bool QPushButton_override_virtual_focusInEvent(void*self, intptr_t slot);
MIQT_EXPORT void QPushButton_virtualbase_focusInEvent(void*self, QFocusEvent*param1);
MIQT_EXPORT bool QPushButton_override_virtual_focusOutEvent(void*self, intptr_t slot);
MIQT_EXPORT void QPushButton_virtualbase_focusOutEvent(void*self, QFocusEvent*param1);
MIQT_EXPORT bool QPushButton_override_virtual_mouseMoveEvent(void*self, intptr_t slot);
MIQT_EXPORT void QPushButton_virtualbase_mouseMoveEvent(void*self, QMouseEvent*param1);
MIQT_EXPORT bool QPushButton_override_virtual_initStyleOption(void*self, intptr_t slot);
MIQT_EXPORT void QPushButton_virtualbase_initStyleOption(void*self, QStyleOptionButton*option);
MIQT_EXPORT bool QPushButton_override_virtual_hitButton(void*self, intptr_t slot);
MIQT_EXPORT bool QPushButton_virtualbase_hitButton(void*self, QPoint*pos);

MIQT_EXPORT void QPushButton_delete(QPushButton*self);

#ifdef __cplusplus
} /* extern C */
#endif

#endif

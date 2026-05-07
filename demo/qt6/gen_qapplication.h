#pragma once
#ifndef MIQT_QT6_GEN_QAPPLICATION_H
#define MIQT_QT6_GEN_QAPPLICATION_H


#include "libmiqt.h"
#include "miqt_export.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
class QApplication;
class QEvent;
class QMetaObject;
class QObject;
class QStyle;
class QWidget;
#else
typedef struct QApplication QApplication;
typedef struct QEvent QEvent;
typedef struct QMetaObject QMetaObject;
typedef struct QObject QObject;
typedef struct QStyle QStyle;
typedef struct QWidget QWidget;
#endif

MIQT_EXPORT QApplication*QApplication_new(int*argc, char** argv);
MIQT_EXPORT QApplication*QApplication_new2(int*argc, char** argv, int param3);
MIQT_EXPORT void QApplication_virtbase(QApplication*src, QObject** outptr_QObject);
MIQT_EXPORT QMetaObject*QApplication_metaObject(QApplication*self);
MIQT_EXPORT void*QApplication_metacast(QApplication*self, char*param1);
MIQT_EXPORT struct miqt_string QApplication_tr(char*s);
MIQT_EXPORT int QApplication_exec();
MIQT_EXPORT void QApplication_setStyle(QApplication*self, QStyle* style);
MIQT_EXPORT QStyle*QApplication_style();
MIQT_EXPORT QWidget*QApplication_focusWidget();
MIQT_EXPORT QWidget*QApplication_activeWindow();
MIQT_EXPORT QWidget*QApplication_widgetAt(int x, int y);
MIQT_EXPORT QWidget*QApplication_topLevelAt(int x, int y);
MIQT_EXPORT void QApplication_closeAllWindows();
MIQT_EXPORT struct miqt_string QApplication_applicationName();

#ifdef __cplusplus
}
#endif

#endif

#pragma once
#ifndef MIQT_QT6_GEN_QTIMER_H
#define MIQT_QT6_GEN_QTIMER_H




#include "libmiqt.h"
#include "miqt_export.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
class QChildEvent;
class QEvent;
class QMetaMethod;
class QMetaObject;
class QObject;
class QTimer;
class QTimerEvent;
using Qt_TimerId = Qt::TimerId;
using Qt_TimerType = Qt::TimerType;
#else
typedef struct QChildEvent QChildEvent;
typedef struct QEvent QEvent;
typedef struct QMetaMethod QMetaMethod;
typedef struct QMetaObject QMetaObject;
typedef struct QObject QObject;
typedef struct QTimer QTimer;
typedef struct QTimerEvent QTimerEvent;
typedef int Qt_TimerId;
typedef int Qt_TimerType;
#endif

MIQT_EXPORT QTimer*QTimer_new();
MIQT_EXPORT QTimer*QTimer_new2(QObject*parent);
MIQT_EXPORT void QTimer_virtbase(QTimer*src, QObject** outptr_QObject);
MIQT_EXPORT QMetaObject*QTimer_metaObject(QTimer*self);
MIQT_EXPORT void*QTimer_metacast(QTimer*self, char*param1);
MIQT_EXPORT struct miqt_string QTimer_tr(char*s);
MIQT_EXPORT bool QTimer_isActive(QTimer*self);
MIQT_EXPORT int QTimer_timerId(QTimer*self);
MIQT_EXPORT Qt_TimerId QTimer_id(QTimer*self);
MIQT_EXPORT void QTimer_setInterval(QTimer*self, int msec);
MIQT_EXPORT int QTimer_interval(QTimer*self);
MIQT_EXPORT int QTimer_remainingTime(QTimer*self);
MIQT_EXPORT void QTimer_setTimerType(QTimer*self, Qt_TimerType atype);
MIQT_EXPORT Qt_TimerType QTimer_timerType(QTimer*self);
MIQT_EXPORT void QTimer_setSingleShot(QTimer*self, bool singleShot);
MIQT_EXPORT bool QTimer_isSingleShot(QTimer*self);
MIQT_EXPORT void QTimer_start(QTimer*self, int msec);
MIQT_EXPORT void QTimer_start2(QTimer*self);
MIQT_EXPORT void QTimer_stop(QTimer*self);
MIQT_EXPORT void QTimer_timerEvent(QTimer*self, QTimerEvent*param1);
MIQT_EXPORT struct miqt_string QTimer_tr2(char*s, char*c);
MIQT_EXPORT struct miqt_string QTimer_tr3(char*s, char*c, int n);

MIQT_EXPORT bool QTimer_override_virtual_timerEvent(void*self, intptr_t slot);
MIQT_EXPORT void QTimer_virtualbase_timerEvent(void*self, QTimerEvent*param1);
MIQT_EXPORT bool QTimer_override_virtual_event(void*self, intptr_t slot);
MIQT_EXPORT bool QTimer_virtualbase_event(void*self, QEvent*event);
MIQT_EXPORT bool QTimer_override_virtual_eventFilter(void*self, intptr_t slot);
MIQT_EXPORT bool QTimer_virtualbase_eventFilter(void*self, QObject*watched, QEvent*event);
MIQT_EXPORT bool QTimer_override_virtual_childEvent(void*self, intptr_t slot);
MIQT_EXPORT void QTimer_virtualbase_childEvent(void*self, QChildEvent*event);
MIQT_EXPORT bool QTimer_override_virtual_customEvent(void*self, intptr_t slot);
MIQT_EXPORT void QTimer_virtualbase_customEvent(void*self, QEvent*event);
MIQT_EXPORT bool QTimer_override_virtual_connectNotify(void*self, intptr_t slot);
MIQT_EXPORT void QTimer_virtualbase_connectNotify(void*self, QMetaMethod*signal);
MIQT_EXPORT bool QTimer_override_virtual_disconnectNotify(void*self, intptr_t slot);
MIQT_EXPORT void QTimer_virtualbase_disconnectNotify(void*self, QMetaMethod*signal);

MIQT_EXPORT QObject*QTimer_protectedbase_sender(bool*_dynamic_cast_ok, void*self);
MIQT_EXPORT int QTimer_protectedbase_senderSignalIndex(bool*_dynamic_cast_ok, void*self);
MIQT_EXPORT int QTimer_protectedbase_receivers(bool*_dynamic_cast_ok, void*self, char*signal);
MIQT_EXPORT bool QTimer_protectedbase_isSignalConnected(bool*_dynamic_cast_ok, void*self, QMetaMethod*signal);

MIQT_EXPORT void QTimer_delete(QTimer*self);

#ifdef __cplusplus
} /* extern C */
#endif

#endif

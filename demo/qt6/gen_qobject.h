#pragma once
#ifndef MIQT_QT6_GEN_QOBJECT_H
#define MIQT_QT6_GEN_QOBJECT_H




#include "libmiqt.h"
#include "miqt_export.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
class QAnyStringView;
class QBindingStorage;
class QChildEvent;
class QEvent;
class QMetaMethod;
class QMetaObject;
#if defined(WORKAROUND_INNER_CLASS_DEFINITION_QMetaObject__Connection)
typedef QMetaObject::Connection QMetaObject__Connection;
#else
class QMetaObject__Connection;
#endif
class QObject;
class QSignalBlocker;
class QThread;
class QTimerEvent;
class QVariant;
using Qt_TimerId = Qt::TimerId;
using Qt_TimerType = Qt::TimerType;
using Qt_ConnectionType = Qt::ConnectionType;
#else
typedef struct QAnyStringView QAnyStringView;
typedef struct QBindingStorage QBindingStorage;
typedef struct QChildEvent QChildEvent;
typedef struct QEvent QEvent;
typedef struct QMetaMethod QMetaMethod;
typedef struct QMetaObject QMetaObject;
typedef struct QMetaObject__Connection QMetaObject__Connection;
typedef struct QObject QObject;
typedef struct QSignalBlocker QSignalBlocker;
typedef struct QThread QThread;
typedef struct QTimerEvent QTimerEvent;
typedef struct QVariant QVariant;
typedef int Qt_TimerId;
typedef int Qt_TimerType;
typedef int Qt_ConnectionType;
#endif

MIQT_EXPORT QObject*QObject_new();
MIQT_EXPORT QObject*QObject_new2(QObject*parent);
MIQT_EXPORT QMetaObject*QObject_metaObject(QObject*self);
MIQT_EXPORT void*QObject_metacast(QObject*self, char*param1);
MIQT_EXPORT struct miqt_string QObject_tr(char*s);
MIQT_EXPORT bool QObject_event(QObject*self, QEvent*event);
MIQT_EXPORT bool QObject_eventFilter(QObject*self, QObject*watched, QEvent*event);
MIQT_EXPORT struct miqt_string QObject_objectName(QObject*self);
MIQT_EXPORT void QObject_setObjectName(QObject*self, QAnyStringView*name);
MIQT_EXPORT bool QObject_isWidgetType(QObject*self);
MIQT_EXPORT bool QObject_isWindowType(QObject*self);
MIQT_EXPORT bool QObject_isQuickItemType(QObject*self);
MIQT_EXPORT bool QObject_isQmlExposed(QObject*self);
MIQT_EXPORT bool QObject_signalsBlocked(QObject*self);
MIQT_EXPORT bool QObject_blockSignals(QObject*self, bool b);
MIQT_EXPORT QThread*QObject_thread(QObject*self);
MIQT_EXPORT bool QObject_moveToThread(QObject*self, QThread*thread);
MIQT_EXPORT int QObject_startTimer(QObject*self, int interval);
MIQT_EXPORT void QObject_killTimer(QObject*self, int id);
MIQT_EXPORT void QObject_killTimerWithId(QObject*self, Qt_TimerId id);
MIQT_EXPORT struct miqt_array /* of QObject**/  QObject_children(QObject*self);
MIQT_EXPORT void QObject_setParent(QObject*self, QObject*parent);
MIQT_EXPORT void QObject_installEventFilter(QObject*self, QObject*filterObj);
MIQT_EXPORT void QObject_removeEventFilter(QObject*self, QObject*obj);
MIQT_EXPORT QMetaObject__Connection*QObject_connect(QObject*sender, QMetaMethod*signal, QObject*receiver, QMetaMethod*method);
MIQT_EXPORT QMetaObject__Connection*QObject_connect2(QObject*self, QObject*sender, char*signal, char*member);
MIQT_EXPORT bool QObject_disconnect(QObject*sender, QMetaMethod*signal, QObject*receiver, QMetaMethod*member);
MIQT_EXPORT bool QObject_disconnectWithQMetaObjectConnection(QMetaObject__Connection*param1);
MIQT_EXPORT void QObject_dumpObjectTree(QObject*self);
MIQT_EXPORT void QObject_dumpObjectInfo(QObject*self);
MIQT_EXPORT bool QObject_setProperty(QObject*self, char*name, QVariant*value);
MIQT_EXPORT QVariant*QObject_property(QObject*self, char*name);
MIQT_EXPORT struct miqt_array /* of struct miqt_string */  QObject_dynamicPropertyNames(QObject*self);
MIQT_EXPORT QBindingStorage*QObject_bindingStorage(QObject*self);
MIQT_EXPORT QBindingStorage*QObject_bindingStorage2(QObject*self);
MIQT_EXPORT void QObject_destroyed(QObject*self);
MIQT_EXPORT void QObject_connect_destroyed(QObject*self, intptr_t slot);
MIQT_EXPORT QObject*QObject_parent(QObject*self);
MIQT_EXPORT bool QObject_inherits(QObject*self, char*classname);
MIQT_EXPORT void QObject_deleteLater(QObject*self);
MIQT_EXPORT void QObject_timerEvent(QObject*self, QTimerEvent*event);
MIQT_EXPORT void QObject_childEvent(QObject*self, QChildEvent*event);
MIQT_EXPORT void QObject_customEvent(QObject*self, QEvent*event);
MIQT_EXPORT void QObject_connectNotify(QObject*self, QMetaMethod*signal);
MIQT_EXPORT void QObject_disconnectNotify(QObject*self, QMetaMethod*signal);
MIQT_EXPORT struct miqt_string QObject_tr2(char*s, char*c);
MIQT_EXPORT struct miqt_string QObject_tr3(char*s, char*c, int n);
MIQT_EXPORT int QObject_startTimer2(QObject*self, int interval, Qt_TimerType timerType);
MIQT_EXPORT QMetaObject__Connection*QObject_connect3(QObject*sender, QMetaMethod*signal, QObject*receiver, QMetaMethod*method, Qt_ConnectionType type);
MIQT_EXPORT QMetaObject__Connection*QObject_connect4(QObject*self, QObject*sender, char*signal, char*member, Qt_ConnectionType type);
MIQT_EXPORT void QObject_destroyedWithQObject(QObject*self, QObject*param1);
MIQT_EXPORT void QObject_connect_destroyedWithQObject(QObject*self, intptr_t slot);

MIQT_EXPORT bool QObject_override_virtual_event(void*self, intptr_t slot);
MIQT_EXPORT bool QObject_virtualbase_event(void*self, QEvent*event);
MIQT_EXPORT bool QObject_override_virtual_eventFilter(void*self, intptr_t slot);
MIQT_EXPORT bool QObject_virtualbase_eventFilter(void*self, QObject*watched, QEvent*event);
MIQT_EXPORT bool QObject_override_virtual_timerEvent(void*self, intptr_t slot);
MIQT_EXPORT void QObject_virtualbase_timerEvent(void*self, QTimerEvent*event);
MIQT_EXPORT bool QObject_override_virtual_childEvent(void*self, intptr_t slot);
MIQT_EXPORT void QObject_virtualbase_childEvent(void*self, QChildEvent*event);
MIQT_EXPORT bool QObject_override_virtual_customEvent(void*self, intptr_t slot);
MIQT_EXPORT void QObject_virtualbase_customEvent(void*self, QEvent*event);
MIQT_EXPORT bool QObject_override_virtual_connectNotify(void*self, intptr_t slot);
MIQT_EXPORT void QObject_virtualbase_connectNotify(void*self, QMetaMethod*signal);
MIQT_EXPORT bool QObject_override_virtual_disconnectNotify(void*self, intptr_t slot);
MIQT_EXPORT void QObject_virtualbase_disconnectNotify(void*self, QMetaMethod*signal);

MIQT_EXPORT QObject*QObject_protectedbase_sender(bool*_dynamic_cast_ok, void*self);
MIQT_EXPORT int QObject_protectedbase_senderSignalIndex(bool*_dynamic_cast_ok, void*self);
MIQT_EXPORT int QObject_protectedbase_receivers(bool*_dynamic_cast_ok, void*self, char*signal);
MIQT_EXPORT bool QObject_protectedbase_isSignalConnected(bool*_dynamic_cast_ok, void*self, QMetaMethod*signal);

MIQT_EXPORT void QObject_delete(QObject*self);

MIQT_EXPORT QSignalBlocker*QSignalBlocker_new(QObject*o);
MIQT_EXPORT QSignalBlocker*QSignalBlocker_new2(QObject*o);
MIQT_EXPORT void QSignalBlocker_reblock(QSignalBlocker*self);
MIQT_EXPORT void QSignalBlocker_unblock(QSignalBlocker*self);
MIQT_EXPORT void QSignalBlocker_dismiss(QSignalBlocker*self);

MIQT_EXPORT void QSignalBlocker_delete(QSignalBlocker*self);

#ifdef __cplusplus
} /* extern C */
#endif

#endif

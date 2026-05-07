#pragma once
#ifndef MIQT_QT6_GEN_QCOREAPPLICATION_H
#define MIQT_QT6_GEN_QCOREAPPLICATION_H




#include "libmiqt.h"
#include "miqt_export.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __cplusplus
class QAbstractEventDispatcher;
class QAbstractNativeEventFilter;
class QChildEvent;
class QCoreApplication;
class QDeadlineTimer;
class QEvent;
#if defined(WORKAROUND_INNER_CLASS_DEFINITION_QEventLoop__ProcessEventsFlags)
typedef QEventLoop::ProcessEventsFlags QEventLoop__ProcessEventsFlags;
#else
class QEventLoop__ProcessEventsFlags;
#endif
class QMetaMethod;
class QMetaObject;
class QObject;
class QPermission;
class QTimerEvent;
class QTranslator;
using Qt_ApplicationAttribute = Qt::ApplicationAttribute;
using Qt_PermissionStatus = Qt::PermissionStatus;
#else
typedef struct QAbstractEventDispatcher QAbstractEventDispatcher;
typedef struct QAbstractNativeEventFilter QAbstractNativeEventFilter;
typedef struct QChildEvent QChildEvent;
typedef struct QCoreApplication QCoreApplication;
typedef struct QDeadlineTimer QDeadlineTimer;
typedef struct QEvent QEvent;
typedef struct QEventLoop__ProcessEventsFlags QEventLoop__ProcessEventsFlags;
typedef struct QMetaMethod QMetaMethod;
typedef struct QMetaObject QMetaObject;
typedef struct QObject QObject;
typedef struct QPermission QPermission;
typedef struct QTimerEvent QTimerEvent;
typedef struct QTranslator QTranslator;
typedef int Qt_ApplicationAttribute;
typedef int Qt_PermissionStatus;
#endif

MIQT_EXPORT QCoreApplication*QCoreApplication_new(int*argc, char** argv);
MIQT_EXPORT QCoreApplication*QCoreApplication_new2(int*argc, char** argv, int param3);
MIQT_EXPORT void QCoreApplication_virtbase(QCoreApplication*src, QObject** outptr_QObject);
MIQT_EXPORT QMetaObject*QCoreApplication_metaObject(QCoreApplication*self);
MIQT_EXPORT void*QCoreApplication_metacast(QCoreApplication*self, char*param1);
MIQT_EXPORT struct miqt_string QCoreApplication_tr(char*s);
MIQT_EXPORT struct miqt_array /* of struct miqt_string */  QCoreApplication_arguments();
MIQT_EXPORT void QCoreApplication_setAttribute(Qt_ApplicationAttribute attribute);
MIQT_EXPORT bool QCoreApplication_testAttribute(Qt_ApplicationAttribute attribute);
MIQT_EXPORT void QCoreApplication_setOrganizationDomain(struct miqt_string orgDomain);
MIQT_EXPORT struct miqt_string QCoreApplication_organizationDomain();
MIQT_EXPORT void QCoreApplication_setOrganizationName(struct miqt_string orgName);
MIQT_EXPORT struct miqt_string QCoreApplication_organizationName();
MIQT_EXPORT void QCoreApplication_setApplicationName(struct miqt_string application);
MIQT_EXPORT struct miqt_string QCoreApplication_applicationName();
MIQT_EXPORT void QCoreApplication_setApplicationVersion(struct miqt_string version);
MIQT_EXPORT struct miqt_string QCoreApplication_applicationVersion();
MIQT_EXPORT void QCoreApplication_setSetuidAllowed(bool allow);
MIQT_EXPORT bool QCoreApplication_isSetuidAllowed();
MIQT_EXPORT QCoreApplication*QCoreApplication_instance();
MIQT_EXPORT bool QCoreApplication_instanceExists();
MIQT_EXPORT int QCoreApplication_exec();
MIQT_EXPORT void QCoreApplication_processEvents();
MIQT_EXPORT void QCoreApplication_processEvents2(QEventLoop__ProcessEventsFlags*flags, int maxtime);
MIQT_EXPORT void QCoreApplication_processEvents3(QEventLoop__ProcessEventsFlags*flags, QDeadlineTimer*deadline);
MIQT_EXPORT bool QCoreApplication_sendEvent(QObject*receiver, QEvent*event);
MIQT_EXPORT void QCoreApplication_postEvent(QObject*receiver, QEvent*event);
MIQT_EXPORT void QCoreApplication_sendPostedEvents();
MIQT_EXPORT void QCoreApplication_removePostedEvents(QObject*receiver);
MIQT_EXPORT QAbstractEventDispatcher*QCoreApplication_eventDispatcher();
MIQT_EXPORT void QCoreApplication_setEventDispatcher(QAbstractEventDispatcher*eventDispatcher);
MIQT_EXPORT bool QCoreApplication_notify(QCoreApplication*self, QObject*param1, QEvent*param2);
MIQT_EXPORT bool QCoreApplication_startingUp();
MIQT_EXPORT bool QCoreApplication_closingDown();
MIQT_EXPORT struct miqt_string QCoreApplication_applicationDirPath();
MIQT_EXPORT struct miqt_string QCoreApplication_applicationFilePath();
MIQT_EXPORT int64_t QCoreApplication_applicationPid();
MIQT_EXPORT Qt_PermissionStatus QCoreApplication_checkPermission(QCoreApplication*self, QPermission*permission);
MIQT_EXPORT void QCoreApplication_setLibraryPaths(struct miqt_array /* of struct miqt_string */  libraryPaths);
MIQT_EXPORT struct miqt_array /* of struct miqt_string */  QCoreApplication_libraryPaths();
MIQT_EXPORT void QCoreApplication_addLibraryPath(struct miqt_string param1);
MIQT_EXPORT void QCoreApplication_removeLibraryPath(struct miqt_string param1);
MIQT_EXPORT bool QCoreApplication_installTranslator(QTranslator*messageFile);
MIQT_EXPORT bool QCoreApplication_removeTranslator(QTranslator*messageFile);
MIQT_EXPORT struct miqt_string QCoreApplication_translate(char*context, char*key);
MIQT_EXPORT void QCoreApplication_installNativeEventFilter(QCoreApplication*self, QAbstractNativeEventFilter*filterObj);
MIQT_EXPORT void QCoreApplication_connect_installNativeEventFilter(QCoreApplication*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_removeNativeEventFilter(QCoreApplication*self, QAbstractNativeEventFilter*filterObj);
MIQT_EXPORT void QCoreApplication_connect_removeNativeEventFilter(QCoreApplication*self, intptr_t slot);
MIQT_EXPORT bool QCoreApplication_isQuitLockEnabled();
MIQT_EXPORT void QCoreApplication_setQuitLockEnabled(bool enabled);
MIQT_EXPORT void QCoreApplication_quit();
MIQT_EXPORT void QCoreApplication_exit();
MIQT_EXPORT void QCoreApplication_organizationNameChanged(QCoreApplication*self);
MIQT_EXPORT void QCoreApplication_connect_organizationNameChanged(QCoreApplication*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_organizationDomainChanged(QCoreApplication*self);
MIQT_EXPORT void QCoreApplication_connect_organizationDomainChanged(QCoreApplication*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_applicationNameChanged(QCoreApplication*self);
MIQT_EXPORT void QCoreApplication_connect_applicationNameChanged(QCoreApplication*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_applicationVersionChanged(QCoreApplication*self);
MIQT_EXPORT void QCoreApplication_connect_applicationVersionChanged(QCoreApplication*self, intptr_t slot);
MIQT_EXPORT bool QCoreApplication_event(QCoreApplication*self, QEvent*param1);
MIQT_EXPORT struct miqt_string QCoreApplication_tr2(char*s, char*c);
MIQT_EXPORT struct miqt_string QCoreApplication_tr3(char*s, char*c, int n);
MIQT_EXPORT void QCoreApplication_setAttribute2(Qt_ApplicationAttribute attribute, bool on);
MIQT_EXPORT void QCoreApplication_processEventsWithFlags(QEventLoop__ProcessEventsFlags*flags);
MIQT_EXPORT void QCoreApplication_postEvent2(QObject*receiver, QEvent*event, int priority);
MIQT_EXPORT void QCoreApplication_sendPostedEventsWithReceiver(QObject*receiver);
MIQT_EXPORT void QCoreApplication_sendPostedEvents2(QObject*receiver, int event_type);
MIQT_EXPORT void QCoreApplication_removePostedEvents2(QObject*receiver, int eventType);
MIQT_EXPORT struct miqt_string QCoreApplication_translate2(char*context, char*key, char*disambiguation);
MIQT_EXPORT struct miqt_string QCoreApplication_translate3(char*context, char*key, char*disambiguation, int n);
MIQT_EXPORT void QCoreApplication_exitWithRetcode(int retcode);

MIQT_EXPORT bool QCoreApplication_override_virtual_notify(void*self, intptr_t slot);
MIQT_EXPORT bool QCoreApplication_virtualbase_notify(void*self, QObject*param1, QEvent*param2);
MIQT_EXPORT bool QCoreApplication_override_virtual_event(void*self, intptr_t slot);
MIQT_EXPORT bool QCoreApplication_virtualbase_event(void*self, QEvent*param1);
MIQT_EXPORT bool QCoreApplication_override_virtual_eventFilter(void*self, intptr_t slot);
MIQT_EXPORT bool QCoreApplication_virtualbase_eventFilter(void*self, QObject*watched, QEvent*event);
MIQT_EXPORT bool QCoreApplication_override_virtual_timerEvent(void*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_virtualbase_timerEvent(void*self, QTimerEvent*event);
MIQT_EXPORT bool QCoreApplication_override_virtual_childEvent(void*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_virtualbase_childEvent(void*self, QChildEvent*event);
MIQT_EXPORT bool QCoreApplication_override_virtual_customEvent(void*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_virtualbase_customEvent(void*self, QEvent*event);
MIQT_EXPORT bool QCoreApplication_override_virtual_connectNotify(void*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_virtualbase_connectNotify(void*self, QMetaMethod*signal);
MIQT_EXPORT bool QCoreApplication_override_virtual_disconnectNotify(void*self, intptr_t slot);
MIQT_EXPORT void QCoreApplication_virtualbase_disconnectNotify(void*self, QMetaMethod*signal);

MIQT_EXPORT void*QCoreApplication_protectedbase_resolveInterface(bool*_dynamic_cast_ok, void*self, char*name, int revision);
MIQT_EXPORT QObject*QCoreApplication_protectedbase_sender(bool*_dynamic_cast_ok, void*self);
MIQT_EXPORT int QCoreApplication_protectedbase_senderSignalIndex(bool*_dynamic_cast_ok, void*self);
MIQT_EXPORT int QCoreApplication_protectedbase_receivers(bool*_dynamic_cast_ok, void*self, char*signal);
MIQT_EXPORT bool QCoreApplication_protectedbase_isSignalConnected(bool*_dynamic_cast_ok, void*self, QMetaMethod*signal);

MIQT_EXPORT void QCoreApplication_delete(QCoreApplication*self);

#ifdef __cplusplus
} /* extern C */
#endif

#endif

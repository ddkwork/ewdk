#include <QApplication>
#include <QWidget>
#include <QStyle>
#include <QMetaObject>
#include "gen_qapplication.h"
#include "libmiqt.h"

#ifdef __cplusplus
extern "C" {
#endif

QApplication* QApplication_new(int* argc, char** argv) {
	return new QApplication(*argc, argv);
}

QApplication* QApplication_new2(int* argc, char** argv, int param3) {
	return new QApplication(*argc, argv, param3);
}

void QApplication_virtbase(QApplication* src, QObject** outptr_QObject) {
	*outptr_QObject = static_cast<QObject*>(src);
}

QMetaObject* QApplication_metaObject(QApplication* self) {
	return (QMetaObject*)self->metaObject();
}

struct miqt_string QApplication_tr(char* s) {
	QString _ret = QApplication::tr(s);
	struct miqt_string _ms;
	_ms.len = _ret.size();
	_ms.data = strdup(_ret.toUtf8().constData());
	return _ms;
}

int QApplication_exec() {
	return QApplication::exec();
}

void QApplication_setStyle(QApplication* self, QStyle* style) {
	self->setStyle(style);
}

QStyle* QApplication_style() {
	return QApplication::style();
}

QWidget* QApplication_focusWidget() {
	return QApplication::focusWidget();
}

QWidget* QApplication_activeWindow() {
	return QApplication::activeWindow();
}

QWidget* QApplication_widgetAt(int x, int y) {
	return QApplication::widgetAt(x, y);
}

QWidget* QApplication_topLevelAt(int x, int y) {
	return QApplication::topLevelAt(x, y);
}

void QApplication_closeAllWindows() {
	QApplication::closeAllWindows();
}

struct miqt_string QApplication_applicationName() {
	QString _ret = QApplication::applicationName();
	struct miqt_string _ms;
	_ms.len = _ret.size();
	_ms.data = strdup(_ret.toUtf8().constData());
	return _ms;
}

#ifdef __cplusplus
}
#endif

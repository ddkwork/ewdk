#include <qevent.h>
#include <qevent.h>
#include <qicon.h>
#include <qevent.h>
#include <qmenu.h>
#include <qmetaobject.h>
#include <qevent.h>
#include <qevent.h>
#include <qpoint.h>
#include <qpushbutton.h>
#include <qsize.h>
#include <qstring.h>
#include <qbytearray.h>
#include <cstring>
#include <qstyleoption.h>
#include <qwidget.h>
#include <qpushbutton.h>
#include "gen_qpushbutton.h"

#ifdef __cplusplus
extern "C" {
#endif

QSize* miqt_exec_callback_QPushButton_sizeHint(const QPushButton*, intptr_t);
QSize* miqt_exec_callback_QPushButton_minimumSizeHint(const QPushButton*, intptr_t);
bool miqt_exec_callback_QPushButton_event(QPushButton*, intptr_t, QEvent*);
void miqt_exec_callback_QPushButton_paintEvent(QPushButton*, intptr_t, QPaintEvent*);
void miqt_exec_callback_QPushButton_keyPressEvent(QPushButton*, intptr_t, QKeyEvent*);
void miqt_exec_callback_QPushButton_focusInEvent(QPushButton*, intptr_t, QFocusEvent*);
void miqt_exec_callback_QPushButton_focusOutEvent(QPushButton*, intptr_t, QFocusEvent*);
void miqt_exec_callback_QPushButton_mouseMoveEvent(QPushButton*, intptr_t, QMouseEvent*);
void miqt_exec_callback_QPushButton_initStyleOption(const QPushButton*, intptr_t, QStyleOptionButton*);
bool miqt_exec_callback_QPushButton_hitButton(const QPushButton*, intptr_t, QPoint*);
#ifdef __cplusplus
} /* extern C */
#endif

class MiqtVirtualQPushButton final : public QPushButton {
public:

	MiqtVirtualQPushButton(QWidget* parent): QPushButton(parent) {}
	MiqtVirtualQPushButton(): QPushButton() {}
	MiqtVirtualQPushButton(const QString& text): QPushButton(text) {}
	MiqtVirtualQPushButton(const QIcon& icon, const QString& text): QPushButton(icon, text) {}
	MiqtVirtualQPushButton(const QString& text, QWidget* parent): QPushButton(text, parent) {}
	MiqtVirtualQPushButton(const QIcon& icon, const QString& text, QWidget* parent): QPushButton(icon, text, parent) {}

	virtual ~MiqtVirtualQPushButton() override = default;

	// cgo.Handle value for overwritten implementation
	intptr_t handle__sizeHint = 0;

	// Subclass to allow providing a Go implementation
	virtual QSize sizeHint() const override {
		if (handle__sizeHint == 0) {
			return QPushButton::sizeHint();
		}

		QSize* callback_return_value = miqt_exec_callback_QPushButton_sizeHint(this, handle__sizeHint);
		return *callback_return_value;
	}

	friend QSize* QPushButton_virtualbase_sizeHint(const void* self);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__minimumSizeHint = 0;

	// Subclass to allow providing a Go implementation
	virtual QSize minimumSizeHint() const override {
		if (handle__minimumSizeHint == 0) {
			return QPushButton::minimumSizeHint();
		}

		QSize* callback_return_value = miqt_exec_callback_QPushButton_minimumSizeHint(this, handle__minimumSizeHint);
		return *callback_return_value;
	}

	friend QSize* QPushButton_virtualbase_minimumSizeHint(const void* self);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__event = 0;

	// Subclass to allow providing a Go implementation
	virtual bool event(QEvent* e) override {
		if (handle__event == 0) {
			return QPushButton::event(e);
		}

		QEvent* sigval1 = e;
		bool callback_return_value = miqt_exec_callback_QPushButton_event(this, handle__event, sigval1);
		return callback_return_value;
	}

	friend bool QPushButton_virtualbase_event(void* self, QEvent* e);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__paintEvent = 0;

	// Subclass to allow providing a Go implementation
	virtual void paintEvent(QPaintEvent* param1) override {
		if (handle__paintEvent == 0) {
			QPushButton::paintEvent(param1);
			return;
		}

		QPaintEvent* sigval1 = param1;
		miqt_exec_callback_QPushButton_paintEvent(this, handle__paintEvent, sigval1);

	}

	friend void QPushButton_virtualbase_paintEvent(void* self, QPaintEvent* param1);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__keyPressEvent = 0;

	// Subclass to allow providing a Go implementation
	virtual void keyPressEvent(QKeyEvent* param1) override {
		if (handle__keyPressEvent == 0) {
			QPushButton::keyPressEvent(param1);
			return;
		}

		QKeyEvent* sigval1 = param1;
		miqt_exec_callback_QPushButton_keyPressEvent(this, handle__keyPressEvent, sigval1);

	}

	friend void QPushButton_virtualbase_keyPressEvent(void* self, QKeyEvent* param1);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__focusInEvent = 0;

	// Subclass to allow providing a Go implementation
	virtual void focusInEvent(QFocusEvent* param1) override {
		if (handle__focusInEvent == 0) {
			QPushButton::focusInEvent(param1);
			return;
		}

		QFocusEvent* sigval1 = param1;
		miqt_exec_callback_QPushButton_focusInEvent(this, handle__focusInEvent, sigval1);

	}

	friend void QPushButton_virtualbase_focusInEvent(void* self, QFocusEvent* param1);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__focusOutEvent = 0;

	// Subclass to allow providing a Go implementation
	virtual void focusOutEvent(QFocusEvent* param1) override {
		if (handle__focusOutEvent == 0) {
			QPushButton::focusOutEvent(param1);
			return;
		}

		QFocusEvent* sigval1 = param1;
		miqt_exec_callback_QPushButton_focusOutEvent(this, handle__focusOutEvent, sigval1);

	}

	friend void QPushButton_virtualbase_focusOutEvent(void* self, QFocusEvent* param1);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__mouseMoveEvent = 0;

	// Subclass to allow providing a Go implementation
	virtual void mouseMoveEvent(QMouseEvent* param1) override {
		if (handle__mouseMoveEvent == 0) {
			QPushButton::mouseMoveEvent(param1);
			return;
		}

		QMouseEvent* sigval1 = param1;
		miqt_exec_callback_QPushButton_mouseMoveEvent(this, handle__mouseMoveEvent, sigval1);

	}

	friend void QPushButton_virtualbase_mouseMoveEvent(void* self, QMouseEvent* param1);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__initStyleOption = 0;

	// Subclass to allow providing a Go implementation
	virtual void initStyleOption(QStyleOptionButton* option) const override {
		if (handle__initStyleOption == 0) {
			QPushButton::initStyleOption(option);
			return;
		}

		QStyleOptionButton* sigval1 = option;
		miqt_exec_callback_QPushButton_initStyleOption(this, handle__initStyleOption, sigval1);

	}

	friend void QPushButton_virtualbase_initStyleOption(const void* self, QStyleOptionButton* option);

	// cgo.Handle value for overwritten implementation
	intptr_t handle__hitButton = 0;

	// Subclass to allow providing a Go implementation
	virtual bool hitButton(const QPoint& pos) const override {
		if (handle__hitButton == 0) {
			return QPushButton::hitButton(pos);
		}

		const QPoint& pos_ret = pos;
		// Cast returned reference into pointer
		QPoint* sigval1 = const_cast<QPoint*>(&pos_ret);
		bool callback_return_value = miqt_exec_callback_QPushButton_hitButton(this, handle__hitButton, sigval1);
		return callback_return_value;
	}

	friend bool QPushButton_virtualbase_hitButton(const void* self, QPoint* pos);

};

QPushButton* QPushButton_new(QWidget* parent) {
	return new (std::nothrow) MiqtVirtualQPushButton(parent);
}

QPushButton* QPushButton_new2() {
	return new (std::nothrow) MiqtVirtualQPushButton();
}

QPushButton* QPushButton_new3(struct miqt_string text) {
	QString text_QString = QString::fromUtf8(text.data, text.len);
	return new (std::nothrow) MiqtVirtualQPushButton(text_QString);
}

QPushButton* QPushButton_new4(QIcon* icon, struct miqt_string text) {
	QString text_QString = QString::fromUtf8(text.data, text.len);
	return new (std::nothrow) MiqtVirtualQPushButton(*icon, text_QString);
}

QPushButton* QPushButton_new5(struct miqt_string text, QWidget* parent) {
	QString text_QString = QString::fromUtf8(text.data, text.len);
	return new (std::nothrow) MiqtVirtualQPushButton(text_QString, parent);
}

QPushButton* QPushButton_new6(QIcon* icon, struct miqt_string text, QWidget* parent) {
	QString text_QString = QString::fromUtf8(text.data, text.len);
	return new (std::nothrow) MiqtVirtualQPushButton(*icon, text_QString, parent);
}

QMetaObject* QPushButton_metaObject(const QPushButton* self) {
	return (QMetaObject*) self->metaObject();
}

void* QPushButton_metacast(QPushButton* self, const char* param1) {
	return self->qt_metacast(param1);
}

struct miqt_string QPushButton_tr(const char* s) {
	QString _ret = QPushButton::tr(s);
	// Convert QString from UTF-16 in C++ RAII memory to UTF-8 in manually-managed C memory
	QByteArray _b = _ret.toUtf8();
	struct miqt_string _ms;
	_ms.len = _b.length();
	_ms.data = static_cast<char*>(malloc(_ms.len));
	memcpy(_ms.data, _b.data(), _ms.len);
	return _ms;
}

QSize* QPushButton_sizeHint(const QPushButton* self) {
	return new QSize(self->sizeHint());
}

QSize* QPushButton_minimumSizeHint(const QPushButton* self) {
	return new QSize(self->minimumSizeHint());
}

bool QPushButton_autoDefault(const QPushButton* self) {
	return self->autoDefault();
}

void QPushButton_setAutoDefault(QPushButton* self, bool autoDefault) {
	self->setAutoDefault(autoDefault);
}

bool QPushButton_isDefault(const QPushButton* self) {
	return self->isDefault();
}

void QPushButton_setDefault(QPushButton* self, bool defaultVal) {
	self->setDefault(defaultVal);
}

void QPushButton_setMenu(QPushButton* self, QMenu* menu) {
	self->setMenu(menu);
}

QMenu* QPushButton_menu(const QPushButton* self) {
	return self->menu();
}

void QPushButton_setFlat(QPushButton* self, bool flat) {
	self->setFlat(flat);
}

bool QPushButton_isFlat(const QPushButton* self) {
	return self->isFlat();
}

void QPushButton_showMenu(QPushButton* self) {
	self->showMenu();
}

struct miqt_string QPushButton_tr2(const char* s, const char* c) {
	QString _ret = QPushButton::tr(s, c);
	// Convert QString from UTF-16 in C++ RAII memory to UTF-8 in manually-managed C memory
	QByteArray _b = _ret.toUtf8();
	struct miqt_string _ms;
	_ms.len = _b.length();
	_ms.data = static_cast<char*>(malloc(_ms.len));
	memcpy(_ms.data, _b.data(), _ms.len);
	return _ms;
}

struct miqt_string QPushButton_tr3(const char* s, const char* c, int n) {
	QString _ret = QPushButton::tr(s, c, static_cast<int>(n));
	// Convert QString from UTF-16 in C++ RAII memory to UTF-8 in manually-managed C memory
	QByteArray _b = _ret.toUtf8();
	struct miqt_string _ms;
	_ms.len = _b.length();
	_ms.data = static_cast<char*>(malloc(_ms.len));
	memcpy(_ms.data, _b.data(), _ms.len);
	return _ms;
}

bool QPushButton_override_virtual_sizeHint(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__sizeHint = slot;
	return true;
}

QSize* QPushButton_virtualbase_sizeHint(const void* self) {
	return new QSize(static_cast<const MiqtVirtualQPushButton*>(self)->QPushButton::sizeHint());
}

bool QPushButton_override_virtual_minimumSizeHint(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__minimumSizeHint = slot;
	return true;
}

QSize* QPushButton_virtualbase_minimumSizeHint(const void* self) {
	return new QSize(static_cast<const MiqtVirtualQPushButton*>(self)->QPushButton::minimumSizeHint());
}

bool QPushButton_override_virtual_event(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__event = slot;
	return true;
}

bool QPushButton_virtualbase_event(void* self, QEvent* e) {
	return static_cast<MiqtVirtualQPushButton*>(self)->QPushButton::event(e);
}

bool QPushButton_override_virtual_paintEvent(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__paintEvent = slot;
	return true;
}

void QPushButton_virtualbase_paintEvent(void* self, QPaintEvent* param1) {
	static_cast<MiqtVirtualQPushButton*>(self)->QPushButton::paintEvent(param1);
}

bool QPushButton_override_virtual_keyPressEvent(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__keyPressEvent = slot;
	return true;
}

void QPushButton_virtualbase_keyPressEvent(void* self, QKeyEvent* param1) {
	static_cast<MiqtVirtualQPushButton*>(self)->QPushButton::keyPressEvent(param1);
}

bool QPushButton_override_virtual_focusInEvent(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__focusInEvent = slot;
	return true;
}

void QPushButton_virtualbase_focusInEvent(void* self, QFocusEvent* param1) {
	static_cast<MiqtVirtualQPushButton*>(self)->QPushButton::focusInEvent(param1);
}

bool QPushButton_override_virtual_focusOutEvent(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__focusOutEvent = slot;
	return true;
}

void QPushButton_virtualbase_focusOutEvent(void* self, QFocusEvent* param1) {
	static_cast<MiqtVirtualQPushButton*>(self)->QPushButton::focusOutEvent(param1);
}

bool QPushButton_override_virtual_mouseMoveEvent(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__mouseMoveEvent = slot;
	return true;
}

void QPushButton_virtualbase_mouseMoveEvent(void* self, QMouseEvent* param1) {
	static_cast<MiqtVirtualQPushButton*>(self)->QPushButton::mouseMoveEvent(param1);
}

bool QPushButton_override_virtual_initStyleOption(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__initStyleOption = slot;
	return true;
}

void QPushButton_virtualbase_initStyleOption(const void* self, QStyleOptionButton* option) {
	static_cast<const MiqtVirtualQPushButton*>(self)->QPushButton::initStyleOption(option);
}

bool QPushButton_override_virtual_hitButton(void* self, intptr_t slot) {
	MiqtVirtualQPushButton* self_cast = static_cast<MiqtVirtualQPushButton*>( (QPushButton*)(self) );
	if (self_cast == nullptr) {
		return false;
	}

	self_cast->handle__hitButton = slot;
	return true;
}

bool QPushButton_virtualbase_hitButton(const void* self, QPoint* pos) {
	return static_cast<const MiqtVirtualQPushButton*>(self)->QPushButton::hitButton(*pos);
}

void QPushButton_delete(QPushButton* self) {
	delete self;
}


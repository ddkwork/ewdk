#include <afxwin.h>

class CMainFrame : public CFrameWnd {
public:
    CMainFrame() {
        Create(nullptr, _T("EWDK MFC Demo"));
    }
};

class CMyApp : public CWinApp {
public:
    virtual BOOL InitInstance() {
        m_pMainWnd = new CMainFrame();
        m_pMainWnd->ShowWindow(SW_SHOW);
        m_pMainWnd->UpdateWindow();
        return TRUE;
    }
};

CMyApp theApp;

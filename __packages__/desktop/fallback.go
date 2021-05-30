//+build !no_fallback

package main

/*
#cgo darwin LDFLAGS: -framework CoreGraphics
#cgo linux pkg-config: x11
#if defined(__APPLE__)
#include <CoreGraphics/CGDisplayConfiguration.h>
int display_width() {
	return CGDisplayPixelsWide(CGMainDisplayID());
}
int display_height() {
	return CGDisplayPixelsHigh(CGMainDisplayID());
}
#elif defined(_WIN32)
#include <wtypes.h>
int display_width() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.right;
}
int display_height() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.bottom;
}
#else
#include <X11/Xlib.h>
int display_width() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->width - 50;
}
int display_height() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->height - 50;
}
#endif
*/
import "C"
import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/webview/webview"
)

func fallback(title string, u string) {
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle(title)
	w.SetSize(int(C.display_width()), int(C.display_height()), 0)
	w.Navigate("data:text/html," + url.PathEscape(fmt.Sprintf(`
	<html>
		<head><title>%s</title></head>
	</html>
	`, title)))

	go func() {
		w.Dispatch(func() {
			for {
				time.Sleep(1 * time.Second)
				_, err := http.Head(u)
				if err == nil {
					break
				}
			}

			w.Navigate(u + "/etabs.html")
		})
	}()

	w.Run()
}

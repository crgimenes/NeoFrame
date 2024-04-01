package screen

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

NSWindow *window;

NSSize getScreenSize() {
	NSRect screenRect = [[NSScreen mainScreen] frame];
	return screenRect.size;
}

void runApp() {
    @autoreleasepool {
		NSSize screenSize = getScreenSize();
        [NSApplication sharedApplication];
        window = [[NSWindow alloc] initWithContentRect:NSMakeRect(
			0,
			0,
			screenSize.width,
			screenSize.height)
		styleMask:NSWindowStyleMaskBorderless
		backing:NSBackingStoreBuffered
		defer:NO];

		[window setOpaque:NO];
		[window setBackgroundColor:[NSColor clearColor]];
		[window setLevel:NSFloatingWindowLevel];
		[window setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces];
		[window setIgnoresMouseEvents:YES];
        [window makeKeyAndOrderFront:nil];
        [NSApp run];
    }
}

void setBackgroudImage(const char *path) {
	@autoreleasepool {
		NSImageView *imageView;
			imageView = [[NSImageView alloc] initWithFrame:[[window contentView] frame]];
			dispatch_async(dispatch_get_main_queue(), ^{
				[[window contentView] addSubview:imageView];
			});

		NSString *imagePath = [NSString stringWithUTF8String:path];
		NSImage *image = [[NSImage alloc] initWithContentsOfFile:imagePath];
		[imageView setImage:image];
	}
}

// set the background image by RGBA array data
void setBackgroudImageByData(unsigned char *data, int width, int height) {
	@autoreleasepool {
		NSImageView *imageView;
			imageView = [[NSImageView alloc] initWithFrame:[[window contentView] frame]];
			dispatch_async(dispatch_get_main_queue(), ^{
				[[window contentView] addSubview:imageView];
			});

		NSBitmapImageRep *bitmap = [[NSBitmapImageRep alloc] initWithBitmapDataPlanes:NULL
			pixelsWide:width
			pixelsHigh:height
			bitsPerSample:8
			samplesPerPixel:4
			hasAlpha:YES
			isPlanar:NO
			colorSpaceName:NSCalibratedRGBColorSpace
			bytesPerRow:width * 4
			bitsPerPixel:32];

		unsigned char *bitmapData = [bitmap bitmapData];
		memcpy(bitmapData, data, width * height * 4);

		NSImage *image = [[NSImage alloc] init];
		[image addRepresentation:bitmap];
		[imageView setImage:image];
	}
}

void clean() {
	@autoreleasepool {
		dispatch_async(dispatch_get_main_queue(), ^{
			[[[window contentView] subviews] makeObjectsPerformSelector:@selector(removeFromSuperview)];
		});
	}
}

*/
import "C"

func SetBackgroudImage(path string) {
	C.setBackgroudImage(C.CString(path))
}

func GetScreenSize() (width, height int) {
	screenSize := C.getScreenSize()
	width = int(screenSize.width)
	height = int(screenSize.height)
	return
}

func SetBackgroudImageByData(data []byte, width, height int) {
	C.setBackgroudImageByData((*C.uchar)(&data[0]), C.int(width), C.int(height))
}

func Clean() {
	C.clean()
}

func RunApp() {
	C.runApp()
}

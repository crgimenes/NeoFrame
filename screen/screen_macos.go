package screen

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

@interface FloatingImageWindow : NSWindow
@end

@implementation FloatingImageWindow

- (BOOL)canBecomeKeyWindow {
    return YES;
}

- (BOOL)canBecomeMainWindow {
    return YES;
}

- (void)mouseDown:(NSEvent *)event {
	// TODO: implementar um modo de sair do programa se o usuário não conseguir ver a tela
	NSLog(@"Mouse down");
	[super mouseDown:event];
}

- (instancetype)initWithContentRect:(NSRect)contentRect {
    self = [super initWithContentRect:contentRect
                            styleMask:NSWindowStyleMaskBorderless
                              backing:NSBackingStoreBuffered
                                defer:NO];
    if (self) {
        [self setOpaque:NO];
        [self setBackgroundColor:[NSColor clearColor]];
        [self setLevel:NSFloatingWindowLevel];
        [self setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces];
		[self setIgnoresMouseEvents:YES]; // TODO: adicionar parametro para que seja opcional
	}
    return self;
}

- (BOOL)isMovableByWindowBackground {
    return YES;
}

@end

FloatingImageWindow *window;

NSSize getScreenSize() {
	NSRect screenRect = [[NSScreen mainScreen] frame];
	return screenRect.size;
}

void showFloatingImageWindow(const char *imageData, int dataSize, int xPosition, int yPosition) {
    NSData *data = [NSData dataWithBytes:imageData length:dataSize];
    NSImage *image = [[NSImage alloc] initWithData:data];
    if (!image) {
        NSLog(@"Failed to load image");
        return;
    }

    //NSSize imageSize = [image size];
	NSSize ws = getScreenSize();
    NSRect windowRect = NSMakeRect(xPosition, yPosition, ws.width, ws.height);

    NSApplication *application = [NSApplication sharedApplication];
    window = [[FloatingImageWindow alloc] initWithContentRect:windowRect];

    NSImageView *imageView = [[NSImageView alloc] initWithFrame:window.contentView.bounds];
    [imageView setImage:image];
    //[imageView setImageScaling:NSImageScaleProportionallyUpOrDown];
	[window setContentView:imageView];

    [window makeKeyAndOrderFront:nil];
    [application run];
}

void changeImage(FloatingImageWindow *window, const char *imageData, int dataSize) {
	NSData *data = [NSData dataWithBytes:imageData length:dataSize];
	NSImage *image = [[NSImage alloc] initWithData:data];
	if (!image) {
		NSLog(@"Failed to load image");
		return;
	}

	NSSize ws = getScreenSize();
	NSRect windowRect = NSMakeRect(
		window.frame.origin.x,
		window.frame.origin.y,
		ws.width,
		ws.height);

	NSImageView *imageView = [[NSImageView alloc] initWithFrame:window.contentView.bounds];
	[imageView setImage:image];
	//[imageView setImageScaling:NSImageScaleProportionallyUpOrDown]
	[window setContentView:imageView];
}

void changeImageOnMainThread(const char *imageData, int dataSize) {
	dispatch_async(dispatch_get_main_queue(), ^{
		changeImage(window, imageData, dataSize);
	});
}

*/
import "C"

func ShowMainWindow(imageData []byte) {
	C.showFloatingImageWindow(
		(*C.char)(C.CBytes(imageData)),
		C.int(len(imageData)),
		C.int(0),
		C.int(0),
	)
}

func ChangeImage(imageData []byte) {
	C.changeImageOnMainThread(
		(*C.char)(C.CBytes(imageData)),
		C.int(len(imageData)),
	)
}

func GetScreenSize() (int, int) {
	screenSize := C.getScreenSize()
	return int(screenSize.width), int(screenSize.height)
}

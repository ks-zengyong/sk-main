/*
* Copyright 2017 Google Inc.
*
* Use of this source code is governed by a BSD-style license that can be
* found in the LICENSE file.
*/

#include "example/HelloWorld.h"

#include "include/core/SkCanvas.h"
#include "include/core/SkColor.h"
#include "include/core/SkFont.h"
#include "include/core/SkFontTypes.h"
#include "include/core/SkGraphics.h"
#include "include/core/SkPaint.h"
#include "include/core/SkPoint.h"
#include "include/core/SkRect.h"
#include "include/core/SkShader.h"
#include "include/core/SkString.h"
#include "include/core/SkSurface.h"
#include "include/core/SkTileMode.h"
#include "include/effects/SkGradientShader.h"
#include "tools/fonts/FontToolUtils.h"
#include "tools/window/DisplayParams.h"

#include <string.h>

using namespace sk_app;
using skwindow::DisplayParams;

Application* Application::Create(int argc, char** argv, void* platformData) {
    return new HelloWorld(argc, argv, platformData);
}

HelloWorld::HelloWorld(int argc, char** argv, void* platformData)
#if defined(SK_GL)
        : fBackendType(Window::kNativeGL_BackendType),
#elif defined(SK_VULKAN)
        : fBackendType(Window::kVulkan_BackendType),
#else
        : fBackendType(Window::kRaster_BackendType),
#endif
        fRotationAngle(0) {
    SkGraphics::Init();

    fWindow = Windows::CreateNativeWindow(platformData);
    fWindow->setRequestedDisplayParams(std::make_unique<DisplayParams>());

    // register callbacks
    fWindow->pushLayer(this);

    fWindow->attach(fBackendType);

    fTypeface = ToolUtils::CreateTypefaceFromResource("fonts/Roboto-Regular.ttf");
    if (!fTypeface) {
        fTypeface = ToolUtils::DefaultPortableTypeface();
    }
}

HelloWorld::~HelloWorld() {
    fWindow->detach();
    delete fWindow;
}

void HelloWorld::updateTitle() {
    if (!fWindow) {
        return;
    }

    SkString title("Hello World ");
    if (Window::kRaster_BackendType == fBackendType) {
        title.append("Raster");
    } else {
#if defined(SK_GL)
        title.append("GL");
#elif defined(SK_VULKAN)
        title.append("Vulkan");
#elif defined(SK_DAWN)
        title.append("Dawn");
#else
        title.append("Unknown GPU backend");
#endif
    }

    fWindow->setTitle(title.c_str());
}

void HelloWorld::onBackendCreated() {
    this->updateTitle();
    fWindow->show();
    fWindow->inval();
}

void drawText(SkCanvas* canvas)
{
    SkFont font = ToolUtils::DefaultPortableFont();
    font.setEdging(SkFont::Edging::kSubpixelAntiAlias);
    //With freetype the default (normal hinting) can be really ugly.
    //Most distros now set slight (vertical hinting only) in any event.
    font.setHinting(SkFontHinting::kSlight);

    const char* text = "Hamburgefons ooo mmm";
    const size_t textLen = strlen(text);

    const SkScalar intervals[] = { 10.0f, 5.0f, 2.0f, 5.0f };
    size_t count = sizeof(intervals) / sizeof(intervals[0]);
    SkPaint paint;
    //paint.setPathEffect(SkDashPathEffect::Make(intervals, count, 0.0f));
    //paint.setStyle(SkPaint::kFill_Style);
    //paint.setStrokeWidth(2.0f);
    paint.setAntiAlias(true);

    SkScalar x = 100;
    SkScalar y = 100;

    SkMatrix mtx;
    mtx.setAll(
        1, 1, 0,
        1, 1, 1,
        -2.0 / 300, 0, 1);
    SkAutoCanvasRestore acr(canvas, true);
    //canvas->translate(SkIntToScalar(50 + 230),SkIntToScalar(20));
    //canvas->rotate(SkIntToScalar(45), x, y * 10);
    canvas->setMatrix(mtx);

    SkRect rc = SkRect::MakeLTRB(0, 0, 100, 100);
    SkRect transRc = mtx.mapRect(rc);
    SkPoint lt = mtx.mapPoint({ 0, 0 });
    SkPoint lb = mtx.mapPoint({ 0, 100 });
    SkPoint rt = mtx.mapPoint({ 100, 0 });
    SkPoint rb = mtx.mapPoint({ 100, 100 });

    //{
    //    SkPaint p;
    //    p.setAntiAlias(true);
    //    SkRect r;
    //    r.setLTRB(0, 0, 100, 100);
    //    canvas->drawRect(r, p);
    //}

    font.setSize(SkIntToScalar(64));
    canvas->drawSimpleText("A", strlen("A"), SkTextEncoding::kUTF8, x, y, font, paint);
}

void HelloWorld::onPaint(SkSurface* surface) {
    auto canvas = surface->getCanvas();

    // Clear background
    canvas->clear(SK_ColorWHITE);

    canvas->save();
    drawText(canvas);
    canvas->restore();

#if 0

    SkPaint paint;
    paint.setColor(SK_ColorRED);

    // Draw a rectangle with red paint
    SkRect rect = SkRect::MakeXYWH(10, 10, 128, 128);
    canvas->drawRect(rect, paint);

    // Set up a linear gradient and draw a circle
    {
        SkPoint linearPoints[] = { { 0, 0 }, { 300, 300 } };
        SkColor linearColors[] = { SK_ColorGREEN, SK_ColorBLACK };
        paint.setShader(SkGradientShader::MakeLinear(linearPoints, linearColors, nullptr, 2,
                                                     SkTileMode::kMirror));
        paint.setAntiAlias(true);

        canvas->drawCircle(200, 200, 64, paint);

        // Detach shader
        paint.setShader(nullptr);
    }

    // Draw a message with a nice black paint
    SkFont font(fTypeface, 20);
    font.setSubpixel(true);
    paint.setColor(SK_ColorBLACK);

    canvas->save();
    static const char message[] = "Hello World ";

    // Translate and rotate
    canvas->translate(300, 300);
    fRotationAngle += 0.2f;
    if (fRotationAngle > 360) {
        fRotationAngle -= 360;
    }
    canvas->rotate(fRotationAngle);

    // Draw the text
    canvas->drawSimpleText(message, strlen(message), SkTextEncoding::kUTF8, 0, 0, font, paint);

    canvas->restore();
#endif
}

void HelloWorld::onIdle() {
    // Just re-paint continuously
    fWindow->inval();
}

bool HelloWorld::onChar(SkUnichar c, skui::ModifierKey modifiers) {
    if (' ' == c) {
        if (Window::kRaster_BackendType == fBackendType) {
#if defined(SK_GL)
            fBackendType = Window::kNativeGL_BackendType;
#elif defined(SK_VULKAN)
            fBackendType = Window::kVulkan_BackendType;
#else
            SkDebugf("No GPU backend configured\n");
            return true;
#endif
        } else {
            fBackendType = Window::kRaster_BackendType;
        }
        fWindow->detach();
        fWindow->attach(fBackendType);
    }
    return true;
}

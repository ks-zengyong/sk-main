/*
 * Copyright 2011 Google Inc.
 *
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

#include "gm/gm.h"
#include "include/core/SkCanvas.h"
#include "include/core/SkFont.h"
#include "include/core/SkFontTypes.h"
#include "include/core/SkPaint.h"
#include "include/core/SkRect.h"
#include "include/core/SkScalar.h"
#include "include/core/SkSize.h"
#include "include/core/SkString.h"
#include "include/core/SkPathEffect.h"
#include "include/effects/SkDashPathEffect.h"
#include "tools/fonts/FontToolUtils.h"

#include <string.h>

namespace skiagm {

class FontScalerGM : public GM {
public:
    FontScalerGM() {
        this->setBGColor(0xFFFFFFFF);
    }

protected:
    SkString getName() const override { return SkString("fontscaler"); }

    SkISize getISize() override { return SkISize::Make(1450, 750); }

    void onDraw(SkCanvas* canvas) override {
        SkFont font = ToolUtils::DefaultPortableFont();
        font.setEdging(SkFont::Edging::kSubpixelAntiAlias);
        //With freetype the default (normal hinting) can be really ugly.
        //Most distros now set slight (vertical hinting only) in any event.
        font.setHinting(SkFontHinting::kSlight);

        const char* text = "Hamburgefons ooo mmm";
        const size_t textLen = strlen(text);

        for (int j = 0; j < 2; ++j) {
            // This used to do 6 iterations but it causes the N4 to crash in the MSAA4 config.
            for (int i = 0; i < 5; ++i) {
                SkScalar x = 10;
                SkScalar y = 20;

                SkAutoCanvasRestore acr(canvas, true);
                canvas->translate(SkIntToScalar(50 + i * 230),
                                  SkIntToScalar(20));
                canvas->rotate(SkIntToScalar(i * 5), x, y * 10);

                {
                    SkPaint p;
                    p.setAntiAlias(true);
                    SkRect r;
                    r.setLTRB(x - 3, 15, x - 1, 280);
                    canvas->drawRect(r, p);
                }

                for (int ps = 6; ps <= 22; ps++) {
                    font.setSize(SkIntToScalar(ps));
                    canvas->drawSimpleText(text, textLen, SkTextEncoding::kUTF8, x, y, font, SkPaint());
                    y += font.getMetrics(nullptr);
                }
            }
            canvas->translate(0, SkIntToScalar(360));
            font.setSubpixel(true);
            font.setLinearMetrics(true);
            font.setBaselineSnap(false);
        }
    }

private:
    using INHERITED = GM;
};


class EffectFontScalerGM : public GM
{
public:
    EffectFontScalerGM()
    {
        this->setBGColor(0xFFFFFFFF);
    }

protected:
    SkString getName() const override { return SkString("effectfontscaler"); }

    SkISize getISize() override { return SkISize::Make(1450, 750); }

    void onDraw(SkCanvas* canvas) override
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
        return;

        //for (int j = 0; j < 2; ++j)
        //{
        //    // This used to do 6 iterations but it causes the N4 to crash in the MSAA4 config.
        //    for (int i = 0; i < 5; ++i)
        //    {
        //        SkScalar x = 10;
        //        SkScalar y = 20;

        //        SkAutoCanvasRestore acr(canvas, true);
        //        canvas->translate(SkIntToScalar(50 + i * 230),
        //            SkIntToScalar(20));
        //        canvas->rotate(SkIntToScalar(i * 5), x, y * 10);

        //        {
        //            SkPaint p;
        //            p.setAntiAlias(true);
        //            SkRect r;
        //            r.setLTRB(x - 3, 15, x - 1, 280);
        //            canvas->drawRect(r, p);
        //        }

        //        for (int ps = 6; ps <= 22; ps++)
        //        {
        //            font.setSize(SkIntToScalar(ps));
        //            canvas->drawSimpleText(text, textLen, SkTextEncoding::kUTF8, x, y, font, paint);
        //            y += font.getMetrics(nullptr);
        //        }
        //    }
        //    canvas->translate(0, SkIntToScalar(360));
        //    font.setSubpixel(true);
        //    font.setLinearMetrics(true);
        //    font.setBaselineSnap(false);
        //}
    }

private:
    using INHERITED = GM;
};

//////////////////////////////////////////////////////////////////////////////

DEF_GM( return new FontScalerGM; )
DEF_GM( return new EffectFontScalerGM; )

}  // namespace skiagm

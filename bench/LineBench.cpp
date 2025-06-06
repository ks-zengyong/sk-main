/*
 * Copyright 2012 Google Inc.
 *
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

#include "bench/Benchmark.h"
#include "include/core/SkBitmap.h"
#include "include/core/SkCanvas.h"
#include "include/core/SkPaint.h"
#include "include/core/SkShader.h"
#include "include/core/SkString.h"
#include "include/private/base/SkTArray.h"
#include "src/base/SkRandom.h"
#include "src/core/SkColorPriv.h"


class LineBench : public Benchmark {
    SkScalar    fStrokeWidth;
    bool        fDoAA;
    SkString    fName;
    enum {
        PTS = 500,
    };
    SkPoint fPts[PTS];

public:
    LineBench(SkScalar width, bool doAA)  {
        fStrokeWidth = width;
        fDoAA = doAA;
        fName.printf("lines_%g_%s", width, doAA ? "AA" : "BW");

        SkRandom rand;
        for (int i = 0; i < PTS; ++i) {
            fPts[i].set(rand.nextUScalar1() * 640, rand.nextUScalar1() * 480);
        }
    }

protected:
    const char* onGetName() override {
        return fName.c_str();
    }

    void onDraw(int loops, SkCanvas* canvas) override {
        SkPaint paint;
        this->setupPaint(&paint);

        paint.setStyle(SkPaint::kStroke_Style);
        paint.setAntiAlias(fDoAA);
        paint.setStrokeWidth(fStrokeWidth);

        for (int i = 0; i < loops; i++) {
            canvas->drawPoints(SkCanvas::kLines_PointMode, PTS, fPts, paint);
        }
    }

private:
    using INHERITED = Benchmark;
};

DEF_BENCH(return new LineBench(0,            false);)
DEF_BENCH(return new LineBench(SK_Scalar1,   false);)
DEF_BENCH(return new LineBench(0,            true);)
DEF_BENCH(return new LineBench(SK_Scalar1/2, true);)
DEF_BENCH(return new LineBench(SK_Scalar1,   true);)

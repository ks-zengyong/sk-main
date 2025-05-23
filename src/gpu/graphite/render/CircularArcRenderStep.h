/*
 * Copyright 2024 Google LLC
 *
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

#ifndef skgpu_graphite_render_CircularArcRenderStep_DEFINED
#define skgpu_graphite_render_CircularArcRenderStep_DEFINED

#include "src/base/SkVx.h"
#include "src/gpu/graphite/Renderer.h"
#include "src/gpu/graphite/ResourceTypes.h"

#include <string>

namespace skgpu::graphite {

class DrawParams;
class DrawWriter;
class PipelineDataGatherer;
class StaticBufferManager;

class CircularArcRenderStep final : public RenderStep {
public:
    CircularArcRenderStep(StaticBufferManager* bufferManager);

    ~CircularArcRenderStep() override;

    std::string vertexSkSL() const override;
    const char* fragmentCoverageSkSL() const override;

    void writeVertices(DrawWriter*, const DrawParams&, skvx::uint2 ssboIndices) const override;
    void writeUniformsAndTextures(const DrawParams&, PipelineDataGatherer*) const override;

private:
    // Points to the static buffers holding the fixed indexed vertex template for drawing instances.
    BindBufferInfo fVertexBuffer;
    BindBufferInfo fIndexBuffer;
};

}  // namespace skgpu::graphite

#endif // skgpu_render_CircularArcRenderStep_DEFINED

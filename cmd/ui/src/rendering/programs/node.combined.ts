// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

/**
 * Sigma.js WebGL Renderer Node Program
 * =====================================
 *
 * Program rendering nodes using GL_TRIANGLE by using an inscribed circle. The
 * program is derived from 'node.image' but overcoming the limitation that
 * gl_PointSize is limited to a specific number of pixels (e.g. macOS: 64px,
 * see https://webglreport.com/?v=2)
 *
 * @module
 */
import { Coordinates, Dimensions, NodeDisplayData } from 'sigma/types';
import { floatColor } from 'sigma/utils';
import { vertexShaderSource } from '../shaders/node.combined.vert';
import { fragmentShaderSource } from '../shaders/node.combined.frag';
import { AbstractNodeProgram } from 'sigma/rendering/webgl/programs/common/node';
import { RenderParams } from 'sigma/rendering/webgl/programs/common/program';
import Sigma from 'sigma/sigma';

const POINTS = 3,
    /*
     atttributes sizing in floats:
      - position (xy: 2xfloat)
      - size (1xfloat)
      - color (4xbyte => 1xfloat)
      - uvw (3xfloat) - only pass width for square texture
      - angle (1xfloat)
      - borderColor (4xbyte = 1xfloat)
   */
    ATTRIBUTES = 9,
    // maximum size of single texture in atlas
    MAX_TEXTURE_SIZE = 192,
    // maximum width of atlas texture (limited by browser)
    // low setting of 2048 works on iPads
    MAX_CANVAS_WIDTH = 2048;

const ANGLE_1 = 0.0,
    ANGLE_2 = (2 * Math.PI) / 3,
    ANGLE_3 = (4 * Math.PI) / 3;

type ImageLoading = { status: 'loading' };
type ImageError = { status: 'error' };
type ImagePending = { status: 'pending'; image: HTMLImageElement };
type ImageReady = { status: 'ready' } & Coordinates & Dimensions;
type ImageType = ImageLoading | ImageError | ImagePending | ImageReady;

// This class only exists for the return typing of `getNodeCombinedProgram`:
class AbstractNodeCombinedProgram extends AbstractNodeProgram {
    constructor(gl: WebGLRenderingContext, renderer: Sigma) {
        super(gl, vertexShaderSource, fragmentShaderSource, POINTS, ATTRIBUTES);
    }
    bind(): void {}
    process(data: NodeDisplayData & { image?: string }, hidden: boolean, offset: number): void {}
    render(params: RenderParams): void {}
    rebindTexture() {}
    /* eslint-enable @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars */
}

/**
 * To share the texture between the program instances of the graph and the
 * hovered nodes (to prevent some flickering, mostly), this program must be
 * "built" for each sigma instance:
 */
export default function getNodeCombinedProgram(): typeof AbstractNodeCombinedProgram {
    /**
     * These attributes are shared between all instances of this exact class,
     * returned by this call to getNodeProgramImage:
     */
    const rebindTextureFns: (() => void)[] = [];
    const images: Record<string, ImageType> = {};
    let textureImage: ImageData;
    let hasReceivedImages = false;
    let pendingImagesFrameID: number | undefined = undefined;

    // next write position in texture
    let writePositionX = 0;
    let writePositionY = 0;
    // height of current row
    let writeRowHeight = 0;

    interface PendingImage {
        image: HTMLImageElement;
        id: string;
        size: number;
    }

    /**
     * Helper to load an image:
     */
    function loadImage(imageSource: string): void {
        if (images[imageSource]) return;

        const image = new Image();
        image.addEventListener('load', () => {
            images[imageSource] = {
                status: 'pending',
                image,
            };

            if (typeof pendingImagesFrameID !== 'number') {
                pendingImagesFrameID = requestAnimationFrame(() => finalizePendingImages());
            }
        });
        image.addEventListener('error', () => {
            images[imageSource] = { status: 'error' };
        });
        images[imageSource] = { status: 'loading' };

        // Load image:
        image.setAttribute('crossOrigin', '');
        image.src = imageSource;
    }

    /**
     * Helper that takes all pending images and adds them into the texture:
     */
    function finalizePendingImages(): void {
        pendingImagesFrameID = undefined;

        const pendingImages: PendingImage[] = [];

        // List all pending images:
        for (const id in images) {
            const state = images[id];
            if (state.status === 'pending') {
                pendingImages.push({
                    id,
                    image: state.image,
                    size: Math.min(state.image.width, state.image.height) || 1,
                });
            }
        }

        // Add images to texture:
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d') as CanvasRenderingContext2D;

        // limit canvas size to avoid browser and platform limits
        let totalWidth = hasReceivedImages ? textureImage.width : 0;
        let totalHeight = hasReceivedImages ? textureImage.height : 0;

        // initialize image drawing offsets with current write position
        let xOffset = writePositionX;
        let yOffset = writePositionY;

        /**
         * Draws a (full or partial) row of images into the atlas texture
         * @param pendingImages
         */
        const drawRow = (pendingImages: PendingImage[]) => {
            // update canvas size before drawing
            if (canvas.width !== totalWidth || canvas.height !== totalHeight) {
                canvas.width = Math.min(MAX_CANVAS_WIDTH, totalWidth);
                canvas.height = totalHeight;

                // draw previous texture into resized canvas
                if (hasReceivedImages) {
                    ctx.putImageData(textureImage, 0, 0);
                }
            }

            pendingImages.forEach(({ id, image, size }) => {
                const imageSizeInTexture = Math.min(MAX_TEXTURE_SIZE, size);

                // Crop image, to only keep the biggest square, centered:
                let dx = 0,
                    dy = 0;
                if ((image.width || 0) > (image.height || 0)) {
                    dx = (image.width - image.height) / 2;
                } else {
                    dy = (image.height - image.width) / 2;
                }

                ctx.drawImage(image, dx, dy, size, size, xOffset, yOffset, imageSizeInTexture, imageSizeInTexture);

                // Update image state:
                images[id] = {
                    status: 'ready',
                    x: xOffset,
                    y: yOffset,
                    width: imageSizeInTexture,
                    height: imageSizeInTexture,
                };

                xOffset += imageSizeInTexture;
            });

            hasReceivedImages = true;
            textureImage = ctx.getImageData(0, 0, canvas.width, canvas.height);
        };

        let rowImages: PendingImage[] = [];
        pendingImages.forEach((image) => {
            const { size } = image;
            const imageSizeInTexture = Math.min(size, MAX_TEXTURE_SIZE);

            if (writePositionX + imageSizeInTexture > MAX_CANVAS_WIDTH) {
                // existing row is full: flush row and continue on next line
                if (rowImages.length > 0) {
                    totalWidth = Math.max(writePositionX, totalWidth);
                    totalHeight = Math.max(writePositionY + writeRowHeight, totalHeight);
                    drawRow(rowImages);

                    rowImages = [];
                    writeRowHeight = 0;
                }

                writePositionX = 0;
                writePositionY = totalHeight;
                xOffset = 0;
                yOffset = totalHeight;
            }

            // add image to row
            rowImages.push(image);

            // advance write position and update maximum row height
            writePositionX += imageSizeInTexture;
            writeRowHeight = Math.max(writeRowHeight, imageSizeInTexture);
        });

        // flush pending images in row - keep write position (and drawing cursor)
        totalWidth = Math.max(writePositionX, totalWidth);
        totalHeight = Math.max(writePositionY + writeRowHeight, totalHeight);
        drawRow(rowImages);
        rowImages = [];

        rebindTextureFns.forEach((fn) => fn());
    }

    return class NodeCombinedProgram extends AbstractNodeProgram {
        texture: WebGLTexture;
        textureLocation: GLint;
        atlasLocation: WebGLUniformLocation;
        sqrtZoomRatioLocation: WebGLUniformLocation;
        correctionRatioLocation: WebGLUniformLocation;
        angleLocation: GLint;
        borderColorLocation: GLint;
        latestRenderParams?: RenderParams;

        constructor(gl: WebGLRenderingContext, renderer: Sigma) {
            super(gl, vertexShaderSource, fragmentShaderSource, POINTS, ATTRIBUTES);

            rebindTextureFns.push(() => {
                if (this && this.rebindTexture) this.rebindTexture();
                if (renderer && renderer.refresh) renderer.refresh();
            });

            textureImage = new ImageData(1, 1);

            // Attribute Location
            this.textureLocation = gl.getAttribLocation(this.program, 'a_texture');
            this.angleLocation = gl.getAttribLocation(this.program, 'a_angle');
            this.borderColorLocation = gl.getAttribLocation(this.program, 'a_borderColor');

            // Uniform Location
            const atlasLocation = gl.getUniformLocation(this.program, 'u_atlas');
            if (atlasLocation === null) throw new Error('NodeProgramImage: error while getting atlasLocation');
            this.atlasLocation = atlasLocation;

            const sqrtZoomRatioLocation = gl.getUniformLocation(this.program, 'u_sqrtZoomRatio');
            if (sqrtZoomRatioLocation === null)
                throw new Error('NodeProgram: error while getting sqrtZoomRatioLocation');
            this.sqrtZoomRatioLocation = sqrtZoomRatioLocation;

            const correctionRatioLocation = gl.getUniformLocation(this.program, 'u_correctionRatio');
            if (correctionRatioLocation === null)
                throw new Error('NodeProgram: error while getting correctionRatioLocation');
            this.correctionRatioLocation = correctionRatioLocation;

            // Initialize WebGL texture:
            this.texture = gl.createTexture() as WebGLTexture;
            gl.bindTexture(gl.TEXTURE_2D, this.texture);
            gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, new Uint8Array([0, 0, 0, 0]));

            // use texture wrapping for debugging only
            //gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
            //gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT);

            this.bind();
        }

        bind(): void {
            super.bind();

            const gl = this.gl;

            gl.enableVertexAttribArray(this.textureLocation);
            gl.enableVertexAttribArray(this.angleLocation);
            gl.enableVertexAttribArray(this.borderColorLocation);

            gl.vertexAttribPointer(
                this.textureLocation,
                3,
                gl.FLOAT,
                false,
                this.attributes * Float32Array.BYTES_PER_ELEMENT,
                16
            );
            gl.vertexAttribPointer(
                this.angleLocation,
                1,
                gl.FLOAT,
                false,
                this.attributes * Float32Array.BYTES_PER_ELEMENT,
                28
            );
            gl.vertexAttribPointer(
                this.borderColorLocation,
                4,
                gl.UNSIGNED_BYTE,
                true,
                this.attributes * Float32Array.BYTES_PER_ELEMENT,
                32
            );
        }

        process(
            data: NodeDisplayData & { image?: string; borderColor?: string },
            hidden: boolean,
            offset: number
        ): void {
            const array = this.array;
            let i = offset * POINTS * ATTRIBUTES;

            const imageSource = data.image;
            const imageState = imageSource && images[imageSource];
            if (typeof imageSource === 'string' && !imageState) loadImage(imageSource);

            if (hidden) {
                // x,y,size,color
                array[i++] = 0;
                array[i++] = 0;
                array[i++] = 0;
                array[i++] = 0;
                // Texture:
                array[i++] = 0;
                array[i++] = 0;
                array[i++] = 0;
                // Angle:
                array[i++] = 0;
                // Border Color:
                array[i++] = 0;
                return;
            }

            const color = floatColor(data.color);
            const borderColor = floatColor(data.borderColor ?? data.color);
            array[i++] = data.x;
            array[i++] = data.y;
            array[i++] = data.size;
            array[i++] = color;

            if (imageState && imageState.status === 'ready') {
                const { width, height } = textureImage;
                // ANGLE_1: center right UV coordinates
                // inscribing circle at (x,y): r=2/3*h, texture (0,0) is top-left
                // texture width is scaled by 2/3 from full triangle width -> uv *1.5
                array[i++] = imageState.x / width + (1.5 * imageState.width) / width;
                array[i++] = imageState.y / height + (0.5 * imageState.height) / height;
                array[i++] = 1;
            } else {
                array[i++] = 0;
                array[i++] = 0;
                array[i++] = 0;
            }
            array[i++] = ANGLE_1;
            array[i++] = borderColor;

            array[i++] = data.x;
            array[i++] = data.y;
            array[i++] = data.size;
            array[i++] = color;

            // ratio: R/texture_height
            const r = (8 / 3) * (1 - Math.sin((2 * Math.PI) / 3));

            if (imageState && imageState.status === 'ready') {
                // ANGLE_2: top left UV coordinates
                const { width, height } = textureImage;
                array[i++] = imageState.x / width;
                array[i++] = imageState.y / height - (r * imageState.height) / height;
                array[i++] = 1;
            } else {
                array[i++] = 0;
                array[i++] = 0;
                array[i++] = 0;
            }
            array[i++] = ANGLE_2;
            array[i++] = borderColor;

            array[i++] = data.x;
            array[i++] = data.y;
            array[i++] = data.size;
            array[i++] = color;
            if (imageState && imageState.status === 'ready') {
                // ANGLE_3: bottom left UV coordinates
                const { width, height } = textureImage;
                array[i++] = imageState.x / width;
                array[i++] = imageState.y / height + (1 + r) * (imageState.height / height);
                array[i++] = 1;
            } else {
                array[i++] = 0;
                array[i++] = 0;
                array[i++] = 0;
            }
            array[i++] = ANGLE_3;
            array[i++] = borderColor;
        }

        render(params: RenderParams): void {
            if (this.hasNothingToRender()) return;

            this.latestRenderParams = params;

            const gl = this.gl;

            const program = this.program;
            gl.useProgram(program);

            gl.uniform1f(this.ratioLocation, 1 / Math.sqrt(params.ratio));
            //gl.uniform1f(this.ratioLocation, 1 / params.ratio);
            gl.uniform1f(this.scaleLocation, params.scalingRatio);
            gl.uniform1f(this.correctionRatioLocation, params.correctionRatio);
            gl.uniform1f(this.sqrtZoomRatioLocation, Math.sqrt(params.ratio));
            //gl.uniform1f(this.sqrtZoomRatioLocation, params.ratio);
            gl.uniformMatrix3fv(this.matrixLocation, false, params.matrix);
            gl.uniform1i(this.atlasLocation, 0);

            gl.drawArrays(gl.TRIANGLES, 0, this.array.length / ATTRIBUTES);
        }

        rebindTexture() {
            const gl = this.gl;

            gl.activeTexture(gl.TEXTURE0);
            gl.bindTexture(gl.TEXTURE_2D, this.texture);
            gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, textureImage);
            gl.generateMipmap(gl.TEXTURE_2D);

            if (this.latestRenderParams) {
                this.bind();
                this.bufferData();
                this.render(this.latestRenderParams);
            }
        }
    };
}

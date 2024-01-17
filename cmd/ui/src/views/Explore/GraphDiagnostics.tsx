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

import { useSigma } from '@react-sigma/core';
import cloneDeep from 'lodash/cloneDeep';
import debounce from 'lodash/debounce';
import { useEffect, useRef, useState } from 'react';

export const GraphDiagnostics = () => {
    const sigma = useSigma();
    const [sigmaState, setSigmaState] = useState(sigma);
    const [mouseCoords, setMouseCoordinates] = useState({ x: 0, y: 0 });
    const boxRef = useRef<any>(null);

    const mouseMoveHandler = (event: MouseEvent) => {
        setMouseCoordinates({
            x: event.clientX,
            y: event.clientY,
        });
    };

    useEffect(() => {
        window.addEventListener('mousemove', mouseMoveHandler);
        return () => {
            window.removeEventListener('mousemove', mouseMoveHandler);
        };
    });

    useEffect(() => {
        const listenerCallback = debounce(() => {
            const newSigma = cloneDeep(sigma);
            setSigmaState(newSigma);
        }, 250);
        sigma.addListener('afterRender', listenerCallback);

        return () => {
            sigma.removeListener('afterRender', listenerCallback);
        };
    });

    if (!sigmaState) {
        return null;
    }

    return (
        <div
            ref={boxRef}
            style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                display: 'flex',
                justifyContent: 'flex-start',
                alignItems: 'center',
                pointerEvents: 'none',
            }}>
            <div
                style={{
                    backgroundColor: 'rgba(255, 255, 255, .9)',
                    margin: '1em',
                    padding: '1em',
                    borderRadius: '4px',
                }}>
                <p style={{ margin: 0, padding: 0 }}>
                    Camera X/Y: {sigmaState.getCamera().x.toFixed(2)}, {sigmaState.getCamera().y.toFixed(2)}
                    <br />
                    Camera Ratio: {sigmaState.getCamera().ratio.toFixed(2)}
                    <br />
                    Camera Angle: {sigmaState.getCamera().angle.toFixed(2)}
                    <br />
                    Graph Dimensions: {sigmaState.getGraphDimensions().width.toFixed(2)} X{' '}
                    {sigmaState.getGraphDimensions().height.toFixed(2)}
                    <br />
                    Graph Ratio:{' '}
                    {(sigmaState.getGraphDimensions().width / sigmaState.getGraphDimensions().height).toFixed(2)}
                    <br />
                    Renderer Dimensions: {sigmaState.getDimensions().width} X {sigmaState.getDimensions().height}
                    <br />
                    Renderer Ratio: {(sigmaState.getDimensions().width / sigmaState.getDimensions().height).toFixed(2)}
                    <br />
                    Framed Graph Rectangle X: {sigmaState.viewRectangle().x1.toFixed(2)},{' '}
                    {sigmaState.viewRectangle().x2.toFixed(2)}
                    <br />
                    Framed Graph Rectangle Y:{' '}
                    {(sigmaState.viewRectangle().y2 - sigmaState.viewRectangle().height).toFixed(2)},{' '}
                    {sigmaState.viewRectangle().y2.toFixed(2)}
                    <br />
                    Framed Graph Ratio:{' '}
                    {(
                        (sigmaState.viewRectangle().x2 - sigmaState.viewRectangle().x1) /
                        (sigmaState.viewRectangle().y2 -
                            (sigmaState.viewRectangle().y2 - sigmaState.viewRectangle().height))
                    ).toFixed(2)}
                    , {}
                    <br />
                    Mouse Coordinates: {mouseCoords.x - boxRef.current?.getBoundingClientRect().x},{' '}
                    {mouseCoords.y - boxRef.current?.getBoundingClientRect().y}
                    <br />
                    Mouse to Graph:{' '}
                    {sigmaState
                        .viewportToGraph({
                            x: mouseCoords.x - boxRef.current?.getBoundingClientRect().x,
                            y: mouseCoords.y - boxRef.current?.getBoundingClientRect().y,
                        })
                        .x.toFixed(2)}
                    ,{' '}
                    {sigmaState
                        .viewportToGraph({
                            x: mouseCoords.x - boxRef.current?.getBoundingClientRect().x,
                            y: mouseCoords.y - boxRef.current?.getBoundingClientRect().y,
                        })
                        .y.toFixed(2)}
                    <br />
                    Mouse to Framed Graph:{' '}
                    {sigmaState
                        .viewportToFramedGraph({
                            x: mouseCoords.x - boxRef.current?.getBoundingClientRect().x,
                            y: mouseCoords.y - boxRef.current?.getBoundingClientRect().y,
                        })
                        .x.toFixed(2)}
                    ,{' '}
                    {sigmaState
                        .viewportToFramedGraph({
                            x: mouseCoords.x - boxRef.current?.getBoundingClientRect().x,
                            y: mouseCoords.y - boxRef.current?.getBoundingClientRect().y,
                        })
                        .y.toFixed(2)}
                    <br />
                </p>
                {/* <p>
                Bounding Box Extents: {sigmaState.getBBox().x}, {sigmaState.getBBox().y}
            </p> */}
            </div>
        </div>
    );
};

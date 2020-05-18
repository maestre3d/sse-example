import React, { useRef, useEffect } from 'react';
import * as d3 from 'd3';

export function Chart(props) {
    const d3DOM = useRef(null)

    useEffect(() => {
        if (props.data && d3DOM.current) {
            const svg = d3.select(d3DOM.current)

            // Bind D3 data
            const update = svg
                .append('g')
                .selectAll('text')
                .data(props.data)

            update.enter()
                .append('text')
                .attr('x', (d, i) => i * 25)
                .attr('y', 40)
                .style('font-size', 24)
                .text((d) => d)

            update
                .attr('x', (d,i) => i* 40)
                .text((d) => d)

            update.exit()
                .remove()
        }

        return () => {

        }
    }, [props.data, d3DOM])

    return (
        <svg
            className='d3-component'
            width={400}
            height={200}
            ref={d3DOM}
        />
    )
}
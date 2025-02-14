import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { FileDiffHunkFields, DiffHunkLineType } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { FileDiffHunks } from './FileDiffHunks'

export const DEMO_HUNKS: FileDiffHunkFields[] = [
    {
        oldRange: { lines: 7, startLine: 3 },
        newRange: { lines: 7, startLine: 3 },
        oldNoNewlineAt: false,
        section: 'func awesomeness(param string) (int, error) {',
        highlight: {
            aborted: false,
            lines: [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    v, err := makeAwesome()',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    if err != nil {',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        fmt.Printf("wow: %v", err)',
                },
                {
                    kind: DiffHunkLineType.DELETED,
                    html: '        return err',
                },
                {
                    kind: DiffHunkLineType.ADDED,
                    html: '        return nil, err',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    }',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '    return v.Score, nil',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '}',
                },
            ],
        },
    },
]

const { add } = storiesOf('web/diffs/FileDiffHunks', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('One diff unified hunk', () => (
    <WebStory>
        {webProps => (
            <FileDiffHunks
                diffMode="unified"
                {...webProps}
                persistLines={boolean('persistLines', true)}
                fileDiffAnchor="abc"
                lineNumbers={boolean('lineNumbers', true)}
                hunks={DEMO_HUNKS}
                className="abcdef"
            />
        )}
    </WebStory>
))

add('One diff split hunk', () => (
    <WebStory>
        {webProps => (
            <FileDiffHunks
                diffMode="split"
                {...webProps}
                persistLines={boolean('persistLines', true)}
                fileDiffAnchor="abc"
                lineNumbers={boolean('lineNumbers', true)}
                hunks={DEMO_HUNKS}
                className="abcdef"
            />
        )}
    </WebStory>
))

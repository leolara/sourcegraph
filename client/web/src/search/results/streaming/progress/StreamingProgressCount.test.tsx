import { mount } from 'enzyme'
import React from 'react'
import { Progress } from '../../../stream'
import { StreamingProgressCount } from './StreamingProgressCount'

describe('StreamingProgressCount', () => {
    it('should render correctly for 0 items in progress', () => {
        const progress: Progress = {
            durationMs: 0,
            matchCount: 0,
            skipped: [],
        }

        expect(mount(<StreamingProgressCount state="loading" progress={progress} />)).toMatchSnapshot()
    })

    it('should render correctly for 1 item complete', () => {
        const progress: Progress = {
            durationMs: 1250,
            matchCount: 1,
            repositoriesCount: 1,
            skipped: [],
        }

        expect(mount(<StreamingProgressCount state="complete" progress={progress} />)).toMatchSnapshot()
    })

    it('should render correctly for 123 items complete', () => {
        const progress: Progress = {
            durationMs: 1250,
            matchCount: 123,
            repositoriesCount: 500,
            skipped: [],
        }

        expect(mount(<StreamingProgressCount state="complete" progress={progress} />)).toMatchSnapshot()
    })

    it('should render correctly for big numbers complete', () => {
        const progress: Progress = {
            durationMs: 52500,
            matchCount: 1234567,
            repositoriesCount: 8901,
            skipped: [],
        }

        expect(mount(<StreamingProgressCount state="complete" progress={progress} />)).toMatchSnapshot()
    })
})

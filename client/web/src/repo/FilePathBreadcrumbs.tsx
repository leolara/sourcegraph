import classNames from 'classnames'
import * as React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { RepoRevision, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { toTreeURL } from '../util/url'

/**
 * Displays a file path in a repository in breadcrumb style, with ancestor path
 * links.
 */
export const FilePathBreadcrumbs: React.FunctionComponent<
    RepoRevision & {
        filePath: string
        isDir: boolean
        repoUrl: string
    }
> = ({ repoName, repoUrl, revision, filePath, isDir }) => {
    const parts = filePath.split('/')
    const partToUrl = (index: number): string => {
        const partPath = parts.slice(0, index + 1).join('/')
        if (isDir || index < parts.length - 1) {
            return toTreeURL({ repoName, revision, filePath: partPath })
        }
        return toPrettyBlobURL({ repoName, revision, filePath: partPath })
    }
    const partToClassName = (index: number): string =>
        index === parts.length - 1 ? 'test-breadcrumb-part-last' : 'part-directory test-breadcrumb-part-directory'

    const spans: JSX.Element[] = [
        <LinkOrSpan
            key="root-dir"
            className="part-directory test-breadcrumb-part-directory"
            to={repoUrl}
            aria-current={false}
        >
            /
        </LinkOrSpan>,
    ]
    for (const [index, part] of parts.entries()) {
        const link = partToUrl(index)
        const className = classNames('part', partToClassName?.(index))
        spans.push(
            <LinkOrSpan
                key={`link-${index}`}
                className={className}
                to={link}
                aria-current={index === parts.length - 1 ? 'page' : 'false'}
            >
                {index < parts.length - 1 ? `${part} /` : part}
            </LinkOrSpan>
        )
    }

    // Important: do not put spaces between the breadcrumbs or spaces will get added when copying the path
    return <span className="file-path-breadcrumbs">{spans}</span>
}

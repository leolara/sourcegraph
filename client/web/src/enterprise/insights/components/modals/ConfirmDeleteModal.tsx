import React from 'react'

import { noop } from 'lodash'

import { Typography, Text } from '@sourcegraph/wildcard'

import { Insight } from '../../core'
import { useDeleteInsight } from '../../hooks/use-delete-insight'

import { ConfirmationModal } from './ConfirmationModal'

type MinimalInsightFields = Pick<Insight, 'title' | 'id' | 'type'>

interface ConfirmDeleteModalProps {
    insight: MinimalInsightFields
    showModal: boolean
    onCancel?: () => void
    onConfirm?: () => void
}

export const ConfirmDeleteModal: React.FunctionComponent<ConfirmDeleteModalProps> = ({
    insight,
    showModal,
    onCancel = noop,
    onConfirm = noop,
}) => {
    const { delete: handleDelete, loading } = useDeleteInsight()

    const handleConfirm = async (): Promise<void> => {
        if (loading) {
            return
        }

        // TODO [VK] Handle error properly in this modal
        await handleDelete(insight)
        onConfirm()
    }

    return (
        <ConfirmationModal
            showModal={showModal}
            onCancel={onCancel}
            onConfirm={handleConfirm}
            ariaLabel="Delete insight modal"
            disabled={loading}
            variant="danger"
            confirmText="Delete forever"
        >
            <Typography.H3 className="text-danger mb-4">Delete '{insight.title}'?</Typography.H3>
            <Text className="mb-4">Are you sure you want to delete insight {insight.title}? This can't be undone.</Text>
        </ConfirmationModal>
    )
}

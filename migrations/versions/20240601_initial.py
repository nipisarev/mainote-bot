"""Initial migration

Revision ID: 20240601_initial
Revises: 
Create Date: 2024-06-01 22:45:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '20240601_initial'
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create user_preferences table
    op.create_table(
        'user_preferences',
        sa.Column('chat_id', sa.Text(), nullable=False),
        sa.Column('notification_time', sa.Text(), nullable=True),
        sa.Column('timezone', sa.Text(), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('CURRENT_TIMESTAMP'), nullable=False),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('CURRENT_TIMESTAMP'), nullable=False),
        sa.PrimaryKeyConstraint('chat_id')
    )


def downgrade() -> None:
    # Drop user_preferences table
    op.drop_table('user_preferences') 
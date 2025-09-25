
from alembic import op
import sqlalchemy as sa


revision = "0001_create_docs_user_schema"
down_revision = None
branch_labels = None
depends_on = None


def upgrade():
    op.create_table(
        "Users",
        sa.Column("id", sa.Integer, primary_key=True, autoincrement=True),
        sa.Column("name", sa.String(255), nullable=False),
        sa.Column("email", sa.String(255), nullable=False, unique=True),
    )

    op.create_table(
        "Documents",
        sa.Column("id", sa.Integer, primary_key=True, autoincrement=True),
        sa.Column("user_id", sa.Integer, sa.ForeignKey("Users.id", ondelete="CASCADE"), nullable=False),
        sa.Column("title", sa.String(255), nullable=False),
    )

    op.create_table(
        "Operations",
        sa.Column("id", sa.Integer, primary_key=True, autoincrement=True),
        sa.Column("document_id", sa.Integer, sa.ForeignKey("Documents.id", ondelete="CASCADE"), nullable=False),
        sa.Column("type", sa.Enum("insert", "delete", name="operation_type"), nullable=False),
        sa.Column("position", sa.Integer, nullable=False),
        sa.Column("text", sa.Text, nullable=False),
        sa.Column("length", sa.Integer, nullable=False),
    )

    op.execute("""
        ALTER TABLE "Operations"
        ADD CONSTRAINT operations_validity CHECK (
            (type = 'insert' AND text <> '' AND length = 0) OR
            (type = 'delete' AND text = '' AND length > 0)
        )
    """)

    op.create_table(
        "DocumentPermissions",
        sa.Column("id", sa.Integer, primary_key=True, autoincrement=True),
        sa.Column("document_id", sa.Integer, sa.ForeignKey("Documents.id", ondelete="CASCADE"), nullable=False),
        sa.Column("user_id", sa.Integer, sa.ForeignKey("Users.id", ondelete="CASCADE"), nullable=False),
        sa.Column("permission", sa.Enum("edit", "view-only", name="permission_type"), nullable=False),
        sa.UniqueConstraint("document_id", "user_id", name="uq_document_user")
    )

    op.execute("""
    CREATE OR REPLACE FUNCTION ensure_owner_permission()
    RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO "DocumentPermissions"(document_id, user_id, permission)
        VALUES (NEW.id, NEW.user_id, 'edit')
        ON CONFLICT (document_id, user_id) DO UPDATE
        SET permission = 'edit';
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

    CREATE TRIGGER trg_ensure_owner_permission
    AFTER INSERT ON "Documents"
    FOR EACH ROW
    EXECUTE FUNCTION ensure_owner_permission();
    """)


def downgrade():
    op.execute("DROP TRIGGER IF EXISTS trg_ensure_owner_permission ON \"Documents\"")
    op.execute("DROP FUNCTION IF EXISTS ensure_owner_permission()")

    op.drop_table("DocumentPermissions")
    op.execute("ALTER TABLE \"Operations\" DROP CONSTRAINT operations_validity")
    op.drop_table("Operations")
    op.drop_table("Documents")
    op.drop_table("Users")

    sa.Enum(name="operation_type").drop(op.get_bind(), checkfirst=True)
    sa.Enum(name="permission_type").drop(op.get_bind(), checkfirst=True)

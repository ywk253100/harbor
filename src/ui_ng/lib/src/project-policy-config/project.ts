export class Project {
    project_id: number;
    owner_id: number;
    name: string;
    creation_time: Date;
    creation_time_str: string;
    deleted: number;
    owner_name: string;
    togglable: boolean;
    update_time: Date;
    current_user_role_id: number;
    repo_count: number;
    has_project_admin_role: boolean;
    is_member: boolean;
    role_name: string;
    metadata: {
      public: string | boolean;
      enable_content_trust: string | boolean;
      prevent_vul: string | boolean;
      severity: string;
      auto_scan: string | boolean;
    };
    constructor () {
        this.metadata.public = false;
        this.metadata.enable_content_trust = false;
        this.metadata.prevent_vul = false;
        this.metadata.severity = 'low';
        this.metadata.auto_scan = false;
    }
}

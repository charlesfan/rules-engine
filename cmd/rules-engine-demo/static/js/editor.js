const { createApp } = Vue;

createApp({
    data() {
        return {
            ruleJSON: '',
            examples: [],
            validationResult: null,
            parsedRuleSet: null
        };
    },
    mounted() {
        this.loadExamples();
    },
    methods: {
        async loadExamples() {
            try {
                const response = await axios.get('/api/rules/examples');
                this.examples = response.data;

                // 預設載入第一個範例
                if (this.examples.length > 0) {
                    this.loadExample(this.examples[0]);
                }
            } catch (error) {
                console.error('Failed to load examples:', error);
            }
        },
        loadExample(example) {
            this.ruleJSON = JSON.stringify(example.rule_set, null, 2);
            this.validateRules();
        },
        async validateRules() {
            try {
                const ruleSet = JSON.parse(this.ruleJSON);
                const response = await axios.post('/api/rules/validate', ruleSet);

                this.validationResult = response.data;
                if (response.data.valid) {
                    this.parsedRuleSet = response.data.rule_set;
                }
            } catch (error) {
                this.validationResult = {
                    valid: false,
                    error: error.response?.data?.error || error.message || 'JSON 格式錯誤'
                };
                this.parsedRuleSet = null;
            }
        },
        openPreview() {
            try {
                const ruleSet = JSON.parse(this.ruleJSON);
                // 儲存到 localStorage 供預覽頁面使用
                localStorage.setItem('preview_rules', this.ruleJSON);
                window.open('/preview', '_blank');
            } catch (error) {
                alert('請先確保規則 JSON 格式正確');
            }
        }
    }
}).mount('#app');

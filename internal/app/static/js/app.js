const { createApp } = Vue

createApp({
    data() {
        return {
            todos: [],
            newTodo: {
                title: '',
                category: 'Work',
                priority: 1
            },
            analytics: {
                total_todos: 0,
                completed_todos: 0,
                completion_rate: 0,
                average_time: 0,
                category_counts: {},
                priority_counts: {}
            },
            charts: {
                completion: null,
                category: null
            }
        }
    },
    computed: {
        sortedTodos() {
            return [...this.todos].sort((a, b) => {
                // 優先度で降順ソート
                if (b.priority !== a.priority) {
                    return b.priority - a.priority;
                }
                // 完了していないものを先に
                if (a.done !== b.done) {
                    return a.done ? 1 : -1;
                }
                // 作成日で降順ソート
                return new Date(b.created_at) - new Date(a.created_at);
            });
        }
    },
    methods: {
        async fetchTodos() {
            try {
                const response = await fetch('/api/todos');
                if (!response.ok) throw new Error('Failed to fetch todos');
                this.todos = await response.json();
                await this.fetchAnalytics();
            } catch (error) {
                console.error('Error fetching todos:', error);
            }
        },

        async fetchAnalytics() {
            try {
                const response = await fetch('/api/analytics');
                if (!response.ok) throw new Error('Failed to fetch analytics');
                this.analytics = await response.json();
                this.updateCharts();
            } catch (error) {
                console.error('Error fetching analytics:', error);
            }
        },

        async addTodo() {
            try {
                const response = await fetch('/api/todos', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(this.newTodo)
                });

                if (!response.ok) throw new Error('Failed to add todo');
                
                // フォームをリセット
                this.newTodo = {
                    title: '',
                    category: 'Work',
                    priority: 1
                };

                // データを再取得
                await this.fetchTodos();
            } catch (error) {
                console.error('Error adding todo:', error);
            }
        },

        async toggleTodo(id) {
            try {
                const response = await fetch(`/api/todos/${id}/toggle`, {
                    method: 'PUT'
                });
                
                if (!response.ok) throw new Error('Failed to toggle todo');
                
                await this.fetchTodos();
            } catch (error) {
                console.error('Error toggling todo:', error);
            }
        },

        formatDate(dateString) {
            if (!dateString) return '';
            const options = { 
                year: 'numeric', 
                month: 'short', 
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            };
            return new Date(dateString).toLocaleDateString(undefined, options);
        },

        getPriorityLabel(priority) {
            switch (priority) {
                case 1: return 'Low';
                case 2: return 'Medium';
                case 3: return 'High';
                default: return 'Unknown';
            }
        },

        updateCharts() {
            // Completion Chart
            const completionCtx = document.getElementById('completionChart');
            if (this.charts.completion) {
                this.charts.completion.destroy();
            }

            this.charts.completion = new Chart(completionCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Completed', 'Pending'],
                    datasets: [{
                        data: [
                            this.analytics.completed_todos,
                            this.analytics.total_todos - this.analytics.completed_todos
                        ],
                        backgroundColor: ['#22c55e', '#e5e7eb']
                    }]
                },
                options: {
                    responsive: true,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        }
                    }
                }
            });

            // Category Chart
            const categoryCtx = document.getElementById('categoryChart');
            if (this.charts.category) {
                this.charts.category.destroy();
            }

            const categoryData = this.analytics.category_counts;
            this.charts.category = new Chart(categoryCtx, {
                type: 'bar',
                data: {
                    labels: Object.keys(categoryData),
                    datasets: [{
                        label: 'Tasks per Category',
                        data: Object.values(categoryData),
                        backgroundColor: '#6366f1'
                    }]
                },
                options: {
                    responsive: true,
                    plugins: {
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                stepSize: 1
                            }
                        }
                    }
                }
            });
        }
    },
    mounted() {
        this.fetchTodos();
    }
}).mount('#app');
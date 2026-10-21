<template>
    <view class="content">
        <navbar back-text="返回" title="page02">
            <view slot="right">向前</view>
        </navbar>
        <view class="form-area">
            <picker @change="bindPickerChange" :value="indexPicker" :range="arrayPicker">
                <view class="picker">
                    当前性别: {{arrayPicker[indexPicker]}}
                </view>
            </picker>
            <view>----------------------------------------</view>
            <radio-group class="radio-group" @change="radioChange">
                <label class="radio" v-for="(item, indexRatio) in itemsRatio" :key="item.name">
                    <radio :value="item.name" :checked="item.checked" />
                    {{item.value}}
                </label>
            </radio-group>
            当前性别: {{indexRatio}}
            <view>----------------------------------------</view>
        </view>
    </view>
</template>

<script>
    export default {
        data() {
            return {
                indexPicker: 0,
                arrayPicker: ['男', '女', '未知'],

                indexRatio: 'BOY',
                itemsRatio: [{
                        name: 'BOY',
                        value: '男',
                        checked: 'true'
                    },
                    {
                        name: 'GIRL',
                        value: '女'
                    },
                    {
                        name: 'OTHER',
                        value: '未知'
                    }
                ]
            }
        },
        onLoad() {

        },
        methods: {
            bindPickerChange: function(e) {
                this.indexPicker = e.detail.value;
            },
            radioChange: function(e) {
                this.indexRatio = e.detail.value;
            }
        }
    }
</script>

<style>
    .content {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
    }

    .form-area {
        margin-top: 200rpx;
        margin-left: auto;
        margin-right: auto;
        margin-bottom: 50rpx;
    }
</style>
#include<bits/stdc++.h>
using namespace std;
#define ll long long int

void solve(){
    int n;
    cin>>n;
    vector<int> arr(n);
    for(int i=0;i<n;i++){
        cin>>arr[i];
    }
    for(auto it:arr){
        cout<<it+5<<" ";
    }
}

int main(){
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t=1;
    solve();
    return 0;
}
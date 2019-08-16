// https://gist.github.com/luki/335bf1195f5c003cd7872f62b67484da

use std::collections::HashMap;
use std::hash::Hash;

pub struct BidirectionalMap<K, V> {
    right_to_left: HashMap<K, V>,
    left_to_right: HashMap<V, K>,
}

impl<K, V> BidirectionalMap<K, V>
where
    K: Hash + Eq + Clone,
    V: Hash + Eq + Clone,
{
    pub fn new() -> BidirectionalMap<K, V> {
        BidirectionalMap {
            right_to_left: HashMap::new(),
            left_to_right: HashMap::new(),
        }
    }
    #[allow(dead_code)]
    pub fn remove_by_key(&mut self, k: K) -> Option<V> {
        match self.right_to_left.remove(&k) {
            Some(v) => {
                let _ = self.left_to_right.remove(&v);
                Some(v)
            }
            None => None,
        }
    }
    #[allow(dead_code)]
    pub fn remove_by_val(&mut self, v: V) -> Option<K> {
        match self.left_to_right.remove(&v) {
            Some(k) => {
                let _ = self.right_to_left.remove(&k);
                Some(k)
            }
            None => None,
        }
    }
    pub fn get_value(&self, key: K) -> Option<&V> {
        self.right_to_left.get(&key)
    }
    pub fn get_key(&self, val: V) -> Option<&K> {
        self.left_to_right.get(&val)
    }
    pub fn insert(&mut self, k: K, v: V) -> Option<V> {
        let _ = self.left_to_right.insert(v.clone(), k.clone());
        self.right_to_left.insert(k, v)
    }
}
